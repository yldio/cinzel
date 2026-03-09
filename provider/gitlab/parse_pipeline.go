// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import (
	"errors"
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/internal/maputil"
	"github.com/zclconf/go-cty/cty"
)

var (
	errUnsupportedBodyType  = errors.New("unsupported body type")
	errUnsupportedBlockBody = errors.New("unsupported block body type")
)

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

		jobMap, err := parseBodyMap(j.Body, hv, "job")
		if err != nil {
			return nil, fmt.Errorf("error in job '%s': %w", j.ID, err)
		}
		jobs[j.ID] = jobMap
	}

	if len(cfg.Workflow) > 1 {
		return nil, errors.New("at most one workflow block is allowed")
	}
	if len(cfg.Workflow) == 1 {
		workflowMap, err := parseBodyMap(cfg.Workflow[0].Body, hv, "workflow")
		if err != nil {
			return nil, fmt.Errorf("error in workflow: %w", err)
		}
		pipeline["workflow"] = workflowMap
	}

	for _, t := range cfg.Templates {
		templateMap, err := parseBodyMap(t.Body, hv, "job")
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
		m, err := parseBodyMap(b.Body, hv, "variable")
		if err != nil {
			return nil, fmt.Errorf("error in variable '%s': %w", b.ID, err)
		}

		nameRaw, hasName := m["name"]
		value, hasValue := m["value"]
		if !hasName || !hasValue {
			return nil, fmt.Errorf("variable '%s' must include 'name' and 'value'", b.ID)
		}

		name, ok := nameRaw.(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("variable '%s' name must be a non-empty string", b.ID)
		}

		description, hasDescription := m["description"]
		if hasDescription {
			result[name] = map[string]any{"value": value, "description": description}
		} else {
			result[name] = value
		}
	}

	return result, nil
}

func parseBodyMap(body hcl.Body, hv *hclparser.HCLVars, scope string) (map[string]any, error) {
	sb, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil, errUnsupportedBodyType
	}

	out := make(map[string]any)
	for _, name := range maputil.SortedKeys(sb.Attributes) {
		attr := sb.Attributes[name]
		switch {
		case scope == "job" && name == "depends_on":
			refs, err := parseReferenceList(attr.Expr, "job")
			if err != nil {
				return nil, err
			}
			arr := make([]any, 0, len(refs))
			for _, ref := range refs {
				arr = append(arr, ref)
			}
			out["needs"] = arr
		default:
			v, err := parseAttr(attr.Expr, hv)
			if err != nil {
				return nil, err
			}
			out[name] = v
		}
	}

	for _, block := range sb.Blocks {
		switch {
		case scope == "job" && block.Type == "rule":
			rule, err := parseBodyMap(block.Body, hv, "rule")
			if err != nil {
				return nil, err
			}
			appendListValue(out, "rules", rule)
		case scope == "workflow" && block.Type == "rule":
			rule, err := parseBodyMap(block.Body, hv, "rule")
			if err != nil {
				return nil, err
			}
			appendListValue(out, "rules", rule)
		case scope == "job" && (block.Type == "artifacts" || block.Type == "cache"):
			child, err := parseBodyMap(block.Body, hv, block.Type)
			if err != nil {
				return nil, err
			}
			out[block.Type] = child
		default:
			child, err := parseBodyMap(block.Body, hv, block.Type)
			if err != nil {
				return nil, err
			}
			addGenericBlock(out, block.Type, block.Labels, child)
		}
	}

	return out, nil
}

func parseAttr(expr hcl.Expression, hv *hclparser.HCLVars) (any, error) {
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

func appendListValue(target map[string]any, key string, value any) {
	list, ok := target[key].([]any)
	if !ok {
		list = []any{}
	}
	list = append(list, value)
	target[key] = list
}

func sortedJobNames(pipeline map[string]any) []string {
	names := make([]string, 0)
	for key, value := range pipeline {
		if _, ok := value.(map[string]any); !ok {
			continue
		}
		switch key {
		case "stages", "variables", "workflow", "default":
			continue
		}
		names = append(names, key)
	}
	sort.Strings(names)
	return names
}
