// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/internal/maputil"
	"github.com/zclconf/go-cty/cty"
)

var errUnsupportedBodyType = errors.New("unsupported body type")

func parseHCLToPipeline(body hcl.Body) (map[string]any, error) {
	var cfg parseConfig
	diags := gohcl.DecodeBody(body, nil, &cfg)

	if diags.HasErrors() {
		return nil, cinzelerror.ProcessHCLDiags(diags)
	}

	hv := hclparser.NewHCLVars()
	pipeline := make(map[string]any)

	if len(cfg.Stages) > 0 {
		stages := make([]any, 0, len(cfg.Stages))

		for _, stage := range cfg.Stages {
			stages = append(stages, stage)
		}
		pipeline["stages"] = stages
	}

	variables, err := parseVariableBlocks(cfg.Variables, hv)
	if err != nil {
		return nil, err
	}

	if len(variables) > 0 {
		pipeline["variables"] = variables
	}

	jobs := make(map[string]any)
	seenJobs := make(map[string]struct{})

	for _, j := range cfg.Jobs {
		if _, exists := seenJobs[j.ID]; exists {
			return nil, fmt.Errorf("duplicate job name '%s'", j.ID)
		}
		seenJobs[j.ID] = struct{}{}

		jobMap, err := parseJobBlock(j, hv)
		if err != nil {
			return nil, fmt.Errorf("error in job '%s': %w", j.ID, err)
		}
		jobs[j.ID] = jobMap
	}

	if len(cfg.Workflow) > 1 {
		return nil, errors.New("at most one workflow block is allowed")
	}

	if len(cfg.Workflow) == 1 {
		workflowMap, err := parseWorkflowBlock(cfg.Workflow[0], hv)
		if err != nil {
			return nil, fmt.Errorf("error in workflow: %w", err)
		}
		pipeline["workflow"] = workflowMap
	}

	if len(cfg.Default) > 1 {
		return nil, errors.New("at most one default block is allowed")
	}

	if len(cfg.Default) == 1 {
		defaultMap, err := parseDefaultBlock(cfg.Default[0], hv)
		if err != nil {
			return nil, fmt.Errorf("error in default: %w", err)
		}
		pipeline["default"] = defaultMap
	}

	if len(cfg.Includes) > 0 {
		includes, err := parseIncludeBlocks(cfg.Includes, hv)
		if err != nil {
			return nil, err
		}

		pipeline["include"] = includes
	}

	for _, t := range cfg.Templates {
		templateMap, err := parseTemplateBlock(t, hv)
		if err != nil {
			return nil, fmt.Errorf("error in template '%s': %w", t.ID, err)
		}
		jobs["."+t.ID] = templateMap
	}

	if err := validatePipeline(pipeline, jobs); err != nil {
		return nil, err
	}

	for name, job := range jobs {
		pipeline[name] = job
	}

	return pipeline, nil
}

func parseVariableBlocks(blocks []hclVariableBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	result := make(map[string]any)

	for _, b := range blocks {
		nameRaw, err := parseAttr(b.Name, hv)
		if err != nil {
			return nil, fmt.Errorf("error in variable '%s': %w", b.ID, err)
		}

		value, err := parseAttr(b.Value, hv)
		if err != nil {
			return nil, fmt.Errorf("error in variable '%s': %w", b.ID, err)
		}

		if nameRaw == nil || value == nil {
			return nil, fmt.Errorf("variable '%s' must include 'name' and 'value'", b.ID)
		}

		name, ok := nameRaw.(string)

		if !ok || name == "" {
			return nil, fmt.Errorf("variable '%s' name must be a non-empty string", b.ID)
		}

		description, err := parseAttr(b.Description, hv)
		if err != nil {
			return nil, fmt.Errorf("error in variable '%s': %w", b.ID, err)
		}

		if description != nil {
			result[name] = map[string]any{"value": value, "description": description}
		} else {
			result[name] = value
		}
	}

	return result, nil
}

func parseWorkflowBlock(block hclWorkflowBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalAttr(out, "name", block.Name, hv); err != nil {
		return nil, err
	}

	rules, err := parseRuleBlocks(block.Rules, hv)
	if err != nil {
		return nil, err
	}

	if len(rules) > 0 {
		out["rules"] = rules
	}

	return out, nil
}

func parseDefaultBlock(block hclDefaultBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalAttr(out, "image", block.Image, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "before_script", block.BeforeScript, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "after_script", block.AfterScript, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "tags", block.Tags, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "interruptible", block.Interruptible, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "retry", block.Retry, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "timeout", block.Timeout, hv); err != nil {
		return nil, err
	}

	if len(block.Cache) > 1 {
		return nil, errors.New("default can include at most one cache block")
	}

	if len(block.Cache) == 1 {
		cache, err := parseCacheBlock(block.Cache[0], hv)
		if err != nil {
			return nil, err
		}

		out["cache"] = cache
	}

	services, err := parseServiceBlocks(block.Services, hv)
	if err != nil {
		return nil, err
	}

	if len(services) > 0 {
		out["services"] = services
	}

	return out, nil
}

func parseJobBlock(block hclJobBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalAttr(out, "stage", block.Stage, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "image", block.Image, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "script", block.Script, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "before_script", block.BeforeScript, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "after_script", block.AfterScript, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "tags", block.Tags, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "when", block.When, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "allow_failure", block.AllowFailure, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "interruptible", block.Interruptible, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "retry", block.Retry, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "timeout", block.Timeout, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "variables", block.Variables, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "environment", block.Environment, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "release", block.Release, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "trigger", block.Trigger, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "parallel", block.Parallel, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "coverage", block.Coverage, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "resource_group", block.ResourceGroup, hv); err != nil {
		return nil, err
	}

	if refs, err := parseReferenceList(block.DependsOn, "job"); err != nil {
		return nil, fmt.Errorf("depends_on: %w", err)
	} else if len(refs) > 0 {
		arr := make([]any, 0, len(refs))

		for _, ref := range refs {
			arr = append(arr, ref)
		}
		out["needs"] = arr
	}

	if refs, err := parseExtendsReferenceList(block.Extends); err != nil {
		return nil, fmt.Errorf("extends: %w", err)
	} else if len(refs) > 0 {
		arr := make([]any, 0, len(refs))

		for _, ref := range refs {
			arr = append(arr, ref)
		}
		out["extends"] = arr
	}

	rules, err := parseRuleBlocks(block.Rules, hv)
	if err != nil {
		return nil, err
	}

	if len(rules) > 0 {
		out["rules"] = rules
	}

	if len(block.Artifacts) > 1 {
		return nil, errors.New("job can include at most one artifacts block")
	}

	if len(block.Artifacts) == 1 {
		artifacts, err := parseArtifactsBlock(block.Artifacts[0], hv)
		if err != nil {
			return nil, err
		}

		out["artifacts"] = artifacts
	}

	if len(block.Cache) > 1 {
		return nil, errors.New("job can include at most one cache block")
	}

	if len(block.Cache) == 1 {
		cache, err := parseCacheBlock(block.Cache[0], hv)
		if err != nil {
			return nil, err
		}

		out["cache"] = cache
	}

	services, err := parseServiceBlocks(block.Services, hv)
	if err != nil {
		return nil, err
	}

	if len(services) > 0 {
		out["services"] = services
	}

	return out, nil
}

func parseTemplateBlock(block hclTemplateBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	job := hclJobBlock{
		ID:            block.ID,
		Stage:         block.Stage,
		Image:         block.Image,
		Script:        block.Script,
		BeforeScript:  block.BeforeScript,
		AfterScript:   block.AfterScript,
		Tags:          block.Tags,
		DependsOn:     block.DependsOn,
		Extends:       block.Extends,
		When:          block.When,
		AllowFailure:  block.AllowFailure,
		Interruptible: block.Interruptible,
		Retry:         block.Retry,
		Timeout:       block.Timeout,
		Variables:     block.Variables,
		Environment:   block.Environment,
		Release:       block.Release,
		Trigger:       block.Trigger,
		Parallel:      block.Parallel,
		Coverage:      block.Coverage,
		ResourceGroup: block.ResourceGroup,
		Rules:         block.Rules,
		Artifacts:     block.Artifacts,
		Cache:         block.Cache,
		Services:      block.Services,
	}

	return parseJobBlock(job, hv)
}

func parseRuleBlocks(blocks []hclRuleBlock, hv *hclparser.HCLVars) ([]any, error) {
	out := make([]any, 0, len(blocks))

	for _, block := range blocks {
		rule := make(map[string]any)

		if err := setOptionalAttr(rule, "if", block.If, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(rule, "when", block.When, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(rule, "allow_failure", block.AllowFailure, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(rule, "changes", block.Changes, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(rule, "exists", block.Exists, hv); err != nil {
			return nil, err
		}

		out = append(out, rule)
	}

	return out, nil
}

func parseArtifactsBlock(block hclArtifactsBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalAttr(out, "paths", block.Paths, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "exclude", block.Exclude, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "expire_in", block.ExpireIn, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "name", block.Name, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "untracked", block.Untracked, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "when", block.When, hv); err != nil {
		return nil, err
	}

	if len(block.Reports) > 1 {
		return nil, errors.New("artifacts can include at most one reports block")
	}

	if len(block.Reports) == 1 {
		reports, err := parseGenericBodyMap(block.Reports[0].Body, hv)
		if err != nil {
			return nil, err
		}

		out["reports"] = reports
	}

	return out, nil
}

func parseCacheBlock(block hclCacheBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalAttr(out, "key", block.Key, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "paths", block.Paths, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "untracked", block.Untracked, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "when", block.When, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "policy", block.Policy, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "fallback_keys", block.FallbackKeys, hv); err != nil {
		return nil, err
	}

	return out, nil
}

func parseServiceBlocks(blocks []hclServiceBlock, hv *hclparser.HCLVars) ([]any, error) {
	out := make([]any, 0, len(blocks))

	for _, block := range blocks {
		service := make(map[string]any)

		if err := setOptionalAttr(service, "name", block.Name, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(service, "alias", block.Alias, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(service, "entrypoint", block.Entrypoint, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(service, "command", block.Command, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(service, "pull_policy", block.PullPolicy, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(service, "variables", block.Variables, hv); err != nil {
			return nil, err
		}

		out = append(out, service)
	}

	return out, nil
}

func parseIncludeBlocks(blocks []hclIncludeBlock, hv *hclparser.HCLVars) (any, error) {
	includes := make([]any, 0, len(blocks))

	for _, block := range blocks {
		include := make(map[string]any)

		if err := setOptionalAttr(include, "local", block.Local, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "project", block.Project, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "file", block.File, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "ref", block.Ref, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "remote", block.Remote, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "template", block.Template, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "component", block.Component, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "inputs", block.Inputs, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		includes = append(includes, include)
	}

	if len(includes) == 1 {
		return includes[0], nil
	}

	return includes, nil
}

func setOptionalAttr(out map[string]any, key string, expr hcl.Expression, hv *hclparser.HCLVars) error {
	value, err := parseAttr(expr, hv)
	if err != nil {
		return err
	}

	if value != nil {
		out[key] = value
	}

	return nil
}

func parseGenericBodyMap(body hcl.Body, hv *hclparser.HCLVars) (map[string]any, error) {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {
		return nil, errUnsupportedBodyType
	}

	out := make(map[string]any)

	for _, name := range maputil.SortedKeys(sb.Attributes) {
		value, err := parseAttr(sb.Attributes[name].Expr, hv)
		if err != nil {
			return nil, err
		}
		out[name] = value
	}

	for _, block := range sb.Blocks {
		child, err := parseGenericBodyMap(block.Body, hv)
		if err != nil {
			return nil, err
		}

		addGenericBlock(out, block.Type, block.Labels, child)
	}

	return out, nil
}

func parseAttr(expr hcl.Expression, hv *hclparser.HCLVars) (any, error) {
	if expr == nil {
		return nil, nil
	}

	hp := hclparser.New(expr, hv)

	if err := hp.Parse(); err != nil {
		return nil, err
	}

	if hp.Result() == cty.NilVal {
		return nil, nil
	}

	return ctyToAny(hp.Result())
}

func parseReferenceList(expr hcl.Expression, expectedRoot string) ([]string, error) {
	if expr == nil {
		return nil, nil
	}

	if isNilOrEmptyCollectionExpr(expr) {
		return nil, nil
	}

	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		ref, err := parseReference(e, expectedRoot)
		if err != nil {
			return nil, err
		}

		return []string{ref}, nil
	case *hclsyntax.TupleConsExpr:
		refs := make([]string, 0, len(e.Exprs))

		for _, item := range e.Exprs {
			traversal, ok := item.(*hclsyntax.ScopeTraversalExpr)

			if !ok {
				return nil, fmt.Errorf("expected a %s reference", expectedRoot)
			}

			ref, err := parseReference(traversal, expectedRoot)
			if err != nil {
				return nil, err
			}

			refs = append(refs, ref)
		}

		return refs, nil
	default:
		return nil, fmt.Errorf("expected %s references", expectedRoot)
	}
}

func parseReference(expr *hclsyntax.ScopeTraversalExpr, expectedRoot string) (string, error) {
	traversal, diags := hcl.AbsTraversalForExpr(expr)

	if diags.HasErrors() {
		return "", cinzelerror.ProcessHCLDiags(diags)
	}

	if len(traversal) < 2 {
		return "", fmt.Errorf("invalid %s reference", expectedRoot)
	}

	root, ok := traversal[0].(hcl.TraverseRoot)

	if !ok || root.Name != expectedRoot {
		return "", fmt.Errorf("invalid reference root, expected '%s'", expectedRoot)
	}

	attr, ok := traversal[1].(hcl.TraverseAttr)

	if !ok {
		return "", fmt.Errorf("invalid %s reference attribute", expectedRoot)
	}

	return attr.Name, nil
}

func parseExtendsReferenceList(expr hcl.Expression) ([]string, error) {
	if expr == nil {
		return nil, nil
	}

	if isNilOrEmptyCollectionExpr(expr) {
		return nil, nil
	}

	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		ref, err := parseExtendsReference(e)
		if err != nil {
			return nil, err
		}

		return []string{ref}, nil
	case *hclsyntax.TupleConsExpr:
		refs := make([]string, 0, len(e.Exprs))

		for _, item := range e.Exprs {
			traversal, ok := item.(*hclsyntax.ScopeTraversalExpr)

			if !ok {
				return nil, fmt.Errorf("expected job/template references for extends")
			}
			ref, err := parseExtendsReference(traversal)
			if err != nil {
				return nil, err
			}
			refs = append(refs, ref)
		}

		return refs, nil
	default:
		return nil, fmt.Errorf("expected job/template references for extends")
	}
}

func parseExtendsReference(expr *hclsyntax.ScopeTraversalExpr) (string, error) {
	traversal, diags := hcl.AbsTraversalForExpr(expr)

	if diags.HasErrors() {
		return "", cinzelerror.ProcessHCLDiags(diags)
	}

	if len(traversal) < 2 {
		return "", fmt.Errorf("invalid extends reference")
	}

	root, ok := traversal[0].(hcl.TraverseRoot)

	if !ok {
		return "", fmt.Errorf("invalid extends reference root")
	}

	attr, ok := traversal[1].(hcl.TraverseAttr)

	if !ok {
		return "", fmt.Errorf("invalid extends reference attribute")
	}

	switch root.Name {
	case "template":
		return "." + attr.Name, nil
	case "job":
		return attr.Name, nil
	default:
		return "", fmt.Errorf("invalid extends reference root, expected 'template' or 'job'")
	}
}

func addGenericBlock(target map[string]any, key string, labels []string, value any) {
	if len(labels) == 1 {
		mapping, ok := target[key].(map[string]any)

		if !ok || mapping == nil {
			mapping = map[string]any{}
		}
		mapping[labels[0]] = value
		target[key] = mapping

		return
	}

	if existing, ok := target[key]; ok {
		switch casted := existing.(type) {
		case []any:
			target[key] = append(casted, value)
		default:
			target[key] = []any{casted, value}
		}

		return
	}

	target[key] = value
}

func isNilOrEmptyCollectionExpr(expr hcl.Expression) bool {
	hp := hclparser.New(expr, hclparser.NewHCLVars())

	if err := hp.Parse(); err != nil {
		return false
	}

	value := hp.Result()

	if value == cty.NilVal {
		return true
	}

	if value.IsNull() {
		return true
	}

	if value.Type().IsTupleType() || value.Type().IsListType() {
		return value.LengthInt() == 0
	}

	return false
}
