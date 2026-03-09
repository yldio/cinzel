// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/internal/maputil"
	"github.com/yldio/cinzel/internal/naming"
	"github.com/yldio/cinzel/internal/yamlwriter"
	"github.com/yldio/cinzel/provider/github/action"
	ghjob "github.com/yldio/cinzel/provider/github/job"
	"github.com/yldio/cinzel/provider/github/step"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
	"github.com/zclconf/go-cty/cty"
)

func parseHCLToWorkflows(body hcl.Body) ([]WorkflowYAMLFile, map[string]any, []ActionYAMLFile, error) {
	var cfg parseConfig
	diags := gohcl.DecodeBody(body, nil, &cfg)

	if diags.HasErrors() {

		return nil, nil, nil, cinzelerror.ProcessHCLDiags(diags)
	}

	hv := hclparser.NewHCLVars()

	if err := cfg.Variables.Parse(hv); err != nil {

		return nil, nil, nil, err
	}

	parsedSteps, err := cfg.Steps.Parse(hv)
	if err != nil {

		return nil, nil, nil, err
	}

	stepMap, err := stepsToMap(parsedSteps)
	if err != nil {

		return nil, nil, nil, err
	}

	parsedJobs := make(map[string]ghjob.Parsed)

	for _, j := range cfg.Jobs {
		jobContent, err := parseJobConfig(j, hv)
		if err != nil {

			return nil, nil, nil, fmt.Errorf("error in job '%s': %w", j.ID, err)
		}

		job := ghjob.NewParsed(j.ID, jobContent)

		if len(job.StepRefs) > 0 {
			steps := make([]any, 0, len(job.StepRefs))

			for _, stepID := range job.StepRefs {
				stepVal, exists := stepMap[stepID]

				if !exists {

					return nil, nil, nil, fmt.Errorf("error in job '%s': cannot find step '%s'", j.ID, stepID)
				}

				steps = append(steps, stepVal)
			}

			job.Body["steps"] = steps
		}

		parsedJobs[j.ID] = job
	}

	if err := validateParsedJobs(parsedJobs); err != nil {

		return nil, nil, nil, err
	}

	parsedWorkflows := make([]WorkflowYAMLFile, 0, len(cfg.Workflows))

	for _, wf := range cfg.Workflows {
		wfContent, err := parseWorkflowConfig(wf, hv)
		if err != nil {

			return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w", wf.ID, err)
		}

		workflow := ghworkflow.NewParsed(wf.ID, wfContent)

		if workflow.Filename == "" {

			return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w", wf.ID, cinzelerror.ErrWorkflowFilenameRequired)
		}

		if err := validateParsedWorkflow(workflow); err != nil {

			return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w", wf.ID, err)
		}

		if len(workflow.JobRefs) > 0 {
			jobs := make(map[string]any)

			for _, jobID := range workflow.JobRefs {
				jobContent, exists := parsedJobs[jobID]

				if !exists {

					return nil, nil, nil, fmt.Errorf("error in workflow '%s': cannot find job '%s'", wf.ID, jobID)
				}

				jobs[jobID] = jobContent.Body
			}

			workflow.Body["jobs"] = jobs
		}

		parsedWorkflows = append(parsedWorkflows, WorkflowYAMLFile{
			Filename: workflow.Filename,
			Content:  workflow.Body,
		})
	}

	parsedActions, err := parseHCLActions(cfg.Actions, hv, stepMap)
	if err != nil {

		return nil, nil, nil, err
	}

	return parsedWorkflows, stepMap, parsedActions, nil
}

func parseJobConfig(cfg hclJobBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalYAMLAttr(out, "name", cfg.Name, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "if", cfg.If, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "uses", cfg.Uses, hv); err != nil {

		return nil, err
	}

	if refs, err := parseReferenceList(cfg.Steps, "step"); err != nil {

		return nil, err
	} else if len(refs) > 0 {
		out["stepsRefs"] = refs
	}

	if refs, err := parseReferenceList(cfg.DependsOn, "job"); err != nil {

		return nil, err
	} else if len(refs) > 0 {
		deps := make([]any, 0, len(refs))

		for _, ref := range refs {
			deps = append(deps, ref)
		}

		out["needs"] = deps
	}

	if err := setOptionalYAMLAttr(out, "secrets", cfg.Secrets, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "continue-on-error", cfg.ContinueOnError, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "timeout-minutes", cfg.TimeoutMinutes, hv); err != nil {

		return nil, err
	}

	for _, usesBlock := range cfg.UsesBlocks {
		usesValue, err := parseUsesBlockFromConfig(usesBlock, hv)
		if err != nil {

			return nil, err
		}

		out["uses"] = usesValue
	}

	for _, block := range cfg.WithBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {

			return nil, err
		}

		withMap := getOrCreateMap(out, "with")
		withMap[key] = value
	}

	for _, block := range cfg.EnvBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {

			return nil, err
		}

		envMap := getOrCreateMap(out, "env")
		envMap[key] = value
	}

	for _, block := range cfg.OutputBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {

			return nil, err
		}

		outputsMap := getOrCreateMap(out, "outputs")
		outputsMap[key] = value
	}

	for _, block := range cfg.SecretBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {

			return nil, err
		}

		secretsMap := getOrCreateMap(out, "secrets")
		secretsMap[key] = value
	}

	for _, block := range cfg.ServiceBlocks {
		serviceVal, err := parseBodyMap(block.Body, hv, "service")
		if err != nil {

			return nil, err
		}

		servicesMap := getOrCreateMap(out, "services")
		servicesMap[block.ID] = serviceVal
	}

	for _, block := range cfg.RunsOnBlocks {
		runsOnValue, err := parseBodyMap(block.Body, hv, "runs_on")
		if err != nil {

			return nil, err
		}

		if runners, ok := runsOnValue["runners"]; ok && len(runsOnValue) == 1 {
			out["runs-on"] = runners
		} else {
			out["runs-on"] = runsOnValue
		}
	}

	for _, block := range cfg.StrategyBlocks {
		strategyValue, err := parseBodyMap(block.Body, hv, "strategy")
		if err != nil {

			return nil, err
		}

		out["strategy"] = strategyValue
	}

	for _, block := range cfg.Permissions {
		child, err := parseBodyMap(block.Body, hv, "permissions")
		if err != nil {

			return nil, err
		}

		out["permissions"] = child
	}

	for _, block := range cfg.Defaults {
		child, err := parseBodyMap(block.Body, hv, "defaults")
		if err != nil {

			return nil, err
		}

		out["defaults"] = child
	}

	for _, block := range cfg.Concurrency {
		child, err := parseBodyMap(block.Body, hv, "concurrency")
		if err != nil {

			return nil, err
		}

		out["concurrency"] = child
	}

	for _, block := range cfg.Container {
		child, err := parseBodyMap(block.Body, hv, "container")
		if err != nil {

			return nil, err
		}

		out["container"] = child
	}

	for _, block := range cfg.Environment {
		child, err := parseBodyMap(block.Body, hv, "environment")
		if err != nil {

			return nil, err
		}

		out["environment"] = child
	}

	return out, nil
}

func parseWorkflowConfig(cfg hclWorkflowBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalYAMLAttr(out, "filename", cfg.Filename, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "name", cfg.Name, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "run-name", cfg.RunName, hv); err != nil {

		return nil, err
	}

	if refs, err := parseReferenceList(cfg.Jobs, "job"); err != nil {

		return nil, err
	} else if len(refs) > 0 {
		out["jobsRefs"] = refs
	}

	if err := setOptionalYAMLAttr(out, "permissions", cfg.Permissions, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "concurrency", cfg.Concurrency, hv); err != nil {

		return nil, err
	}

	for _, on := range cfg.On {
		eventValue, err := parseBodyMap(on.Body, hv, "on")
		if err != nil {

			return nil, err
		}

		eventName := on.ID
		eventValue = ghworkflow.NormalizeOnEvent(eventName, eventValue)

		onMap := getOrCreateMap(out, "on")
		if eventName == "schedule" {
			onMap[eventName] = ghworkflow.DenormalizeScheduleEvent(eventValue)
		} else if len(eventValue) == 0 {
			onMap[eventName] = map[string]any{}
		} else {
			onMap[eventName] = eventValue
		}
	}

	for _, block := range cfg.Env {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {

			return nil, err
		}

		envMap := getOrCreateMap(out, "env")
		envMap[key] = value
	}

	for _, block := range cfg.PermBlocks {
		child, err := parseBodyMap(block.Body, hv, "permissions")
		if err != nil {

			return nil, err
		}

		out["permissions"] = child
	}

	for _, block := range cfg.Defaults {
		child, err := parseBodyMap(block.Body, hv, "defaults")
		if err != nil {

			return nil, err
		}

		out["defaults"] = child
	}

	for _, block := range cfg.ConcBlocks {
		child, err := parseBodyMap(block.Body, hv, "concurrency")
		if err != nil {

			return nil, err
		}

		out["concurrency"] = child
	}

	return out, nil
}

func setOptionalYAMLAttr(out map[string]any, yamlKey string, expr hcl.Expression, hv *hclparser.HCLVars) error {
	val, err := parseAttr(expr, hv)
	if err != nil {

		return err
	}

	if val != nil {
		out[yamlKey] = val
	}

	return nil
}

func parseNamedConfig(cfg hclNamedBlock, hv *hclparser.HCLVars) (string, any, error) {
	rawName, err := parseAttr(cfg.Name, hv)
	if err != nil {

		return "", nil, err
	}

	name, ok := rawName.(string)

	if !ok || name == "" {

		return "", nil, errors.New("'name' attribute must be a non-empty string")
	}

	value, err := parseAttr(cfg.Value, hv)
	if err != nil {

		return "", nil, err
	}

	return name, value, nil
}

func parseUsesBlockFromConfig(cfg hclUsesBlock, hv *hclparser.HCLVars) (string, error) {
	list := action.UsesListConfig{{Action: cfg.Action, Version: cfg.Version}}
	val, err := list.Parse(hv)
	if err != nil {

		return "", err
	}

	return val.AsString(), nil
}

func parseBodyMap(body hcl.Body, hv *hclparser.HCLVars, scope string) (map[string]any, error) {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {

		return nil, errUnsupportedBodyType
	}

	out := make(map[string]any)

	attrNames := maputil.SortedKeys(sb.Attributes)

	for _, name := range attrNames {
		attr := sb.Attributes[name]

		switch {
		case scope == "workflow" && name == "jobs":
			refs, err := parseReferenceList(attr.Expr, "job")
			if err != nil {

				return nil, err
			}
			out["jobsRefs"] = refs
		case scope == "job" && name == "steps":
			refs, err := parseReferenceList(attr.Expr, "step")
			if err != nil {

				return nil, err
			}
			out["stepsRefs"] = refs
		case scope == "job" && name == "depends_on":
			refs, err := parseReferenceList(attr.Expr, "job")
			if err != nil {

				return nil, err
			}

			deps := make([]any, 0, len(refs))

			for _, ref := range refs {
				deps = append(deps, ref)
			}

			out["needs"] = deps
		default:
			val, err := parseAttr(attr.Expr, hv)
			if err != nil {

				return nil, err
			}

			out[naming.ToYAMLKey(name)] = val
		}
	}

	for _, block := range sb.Blocks {
		switch {
		case scope == "workflow" && block.Type == "on":
			if len(block.Labels) != 1 {

				return nil, errors.New("on block must have exactly one label")
			}

			onMap := getOrCreateMap(out, "on")
			eventValue, err := parseBodyMap(block.Body, hv, "on")
			if err != nil {

				return nil, err
			}

			eventName := block.Labels[0]
			eventValue = ghworkflow.NormalizeOnEvent(eventName, eventValue)

			if eventName == "schedule" {
				onMap[eventName] = ghworkflow.DenormalizeScheduleEvent(eventValue)
			} else if len(eventValue) == 0 {
				onMap[eventName] = map[string]any{}
			} else {
				onMap[eventName] = eventValue
			}
		case block.Type == "uses":
			usesValue, err := parseUsesBlock(block.Body, hv)
			if err != nil {

				return nil, err
			}

			out["uses"] = usesValue
		case block.Type == "with":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {

				return nil, err
			}

			withMap := getOrCreateMap(out, "with")
			withMap[key] = value
		case block.Type == "env":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {

				return nil, err
			}

			envMap := getOrCreateMap(out, "env")
			envMap[key] = value
		case block.Type == "output" && scope == "job":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {

				return nil, err
			}

			outputsMap := getOrCreateMap(out, "outputs")
			outputsMap[key] = value
		case block.Type == "secret" && scope == "job":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {

				return nil, err
			}

			secretsMap := getOrCreateMap(out, "secrets")
			secretsMap[key] = value
		case block.Type == "service" && scope == "job":
			if len(block.Labels) != 1 {

				return nil, errors.New("service block must have exactly one label")
			}

			serviceVal, err := parseBodyMap(block.Body, hv, "service")
			if err != nil {

				return nil, err
			}

			servicesMap := getOrCreateMap(out, "services")
			servicesMap[block.Labels[0]] = serviceVal
		case block.Type == "runs_on" && scope == "job":
			runsOnValue, err := parseBodyMap(block.Body, hv, "runs_on")
			if err != nil {

				return nil, err
			}

			if runners, ok := runsOnValue["runners"]; ok && len(runsOnValue) == 1 {
				out["runs-on"] = runners
			} else {
				out["runs-on"] = runsOnValue
			}
		case block.Type == "matrix" && scope == "strategy":
			matrixValue, err := parseBodyMap(block.Body, hv, "matrix")
			if err != nil {

				return nil, err
			}

			normalized, err := ghjob.NormalizeStrategyMatrix(matrixValue)
			if err != nil {

				return nil, err
			}

			out["matrix"] = normalized
		default:
			child, err := parseBodyMap(block.Body, hv, block.Type)
			if err != nil {

				return nil, err
			}

			addGenericBlock(out, naming.ToYAMLKey(block.Type), block.Labels, child)
		}
	}

	return out, nil
}

func parseUsesBlock(body hcl.Body, hv *hclparser.HCLVars) (string, error) {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {

		return "", errUnsupportedUsesBody
	}

	var cfg action.UsesConfig

	if attr, ok := sb.Attributes["action"]; ok {
		cfg.Action = attr.Expr
	}

	if attr, ok := sb.Attributes["version"]; ok {
		cfg.Version = attr.Expr
	}

	list := action.UsesListConfig{cfg}
	val, err := list.Parse(hv)
	if err != nil {

		return "", err
	}

	return val.AsString(), nil
}

func parseNamedBlock(body hcl.Body, hv *hclparser.HCLVars) (string, any, error) {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {

		return "", nil, errUnsupportedBlockBody
	}

	nameAttr, ok := sb.Attributes["name"]

	if !ok {

		return "", nil, errNamedBlockMissingName
	}

	valueAttr, ok := sb.Attributes["value"]

	if !ok {

		return "", nil, errNamedBlockMissingValue
	}

	rawName, err := parseAttr(nameAttr.Expr, hv)
	if err != nil {

		return "", nil, err
	}

	name, ok := rawName.(string)

	if !ok || name == "" {

		return "", nil, errors.New("'name' attribute must be a non-empty string")
	}

	value, err := parseAttr(valueAttr.Expr, hv)
	if err != nil {

		return "", nil, err
	}

	return name, value, nil
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

func getOrCreateMap(target map[string]any, key string) map[string]any {
	existing, ok := target[key]

	if ok {
		mapping, castOK := existing.(map[string]any)

		if castOK {

			return mapping
		}
	}

	mapping := map[string]any{}
	target[key] = mapping

	return mapping
}

func stepsToMap(steps step.Steps) (map[string]any, error) {
	out := make(map[string]any, len(steps))

	for stepID, parsedStep := range steps {
		converted, err := yamlwriter.Convert(parsedStep)
		if err != nil {

			return nil, err
		}

		out[stepID] = converted
	}

	return out, nil
}
