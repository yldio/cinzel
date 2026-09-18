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

func parseHCLToPipeline(body hcl.Body, sources map[string][]byte) (map[string]any, *comments, error) {
	var cfg parseConfig
	diags := gohcl.DecodeBody(body, nil, &cfg)

	if diags.HasErrors() {
		return nil, nil, cinzelerror.ProcessHCLDiags(diags)
	}

	hv := hclparser.NewHCLVars()
	hv.SetSources(sources)
	pipeline := make(map[string]any)

	// GitLab still reads these five at the top level, where they mean what the
	// same keys mean under "default".
	for _, attr := range [...]struct {
		name string
		expr hcl.Expression
	}{
		{"image", cfg.Image},
		{"before_script", cfg.BeforeScript},
		{"after_script", cfg.AfterScript},
		{"cache", cfg.Cache},
		{"services", cfg.Services},
	} {
		if err := setOptionalAttr(pipeline, attr.name, attr.expr, hv); err != nil {
			return nil, nil, err
		}
	}

	if len(cfg.Stages) > 0 {
		stages := make([]any, 0, len(cfg.Stages))

		for _, stage := range cfg.Stages {
			stages = append(stages, stage)
		}
		pipeline["stages"] = stages
	}

	// A variable block is labelled but is written out under the value of its
	// "name" attribute, so the two have to be matched up before its comments
	// can be filed under the right key.
	variableNames := map[string]string{}

	variables, err := parseVariableBlocks(cfg.Variables, variableNames, hv)
	if err != nil {
		return nil, nil, err
	}

	if len(variables) > 0 {
		pipeline["variables"] = variables
	}

	jobs := make(map[string]any)
	seenJobs := make(map[string]struct{})
	// keys maps a block label to the YAML key the block is written under. The two
	// differ when the block carries an "id" attribute; "needs" and "extends" name
	// labels and have to be rewritten once every key is known.
	keys := make(map[string]string)

	for _, j := range cfg.Jobs {
		if _, exists := seenJobs[j.ID]; exists {
			return nil, nil, fmt.Errorf("duplicate job name '%s'", j.ID)
		}
		seenJobs[j.ID] = struct{}{}

		jobMap, err := parseJobBlock(j, hv)
		if err != nil {
			return nil, nil, fmt.Errorf("error in job '%s': %w", j.ID, err)
		}

		key, err := blockKey(j.ID, j.Key, hv)
		if err != nil {
			return nil, nil, fmt.Errorf("error in job '%s': %w", j.ID, err)
		}

		if _, taken := jobs[key]; taken {
			return nil, nil, fmt.Errorf("duplicate job name '%s'", key)
		}

		keys[j.ID] = key
		jobs[key] = jobMap
	}

	if len(cfg.Workflow) > 1 {
		return nil, nil, errors.New("at most one workflow block is allowed")
	}

	if len(cfg.Workflow) == 1 {
		workflowMap, err := parseWorkflowBlock(cfg.Workflow[0], hv)
		if err != nil {
			return nil, nil, fmt.Errorf("error in workflow: %w", err)
		}
		pipeline["workflow"] = workflowMap
	}

	if len(cfg.Default) > 1 {
		return nil, nil, errors.New("at most one default block is allowed")
	}

	if len(cfg.Default) == 1 {
		defaultMap, err := parseDefaultBlock(cfg.Default[0], hv)
		if err != nil {
			return nil, nil, fmt.Errorf("error in default: %w", err)
		}
		pipeline["default"] = defaultMap
	}

	if len(cfg.Spec) > 1 {
		return nil, nil, errors.New("at most one spec block is allowed")
	}

	if len(cfg.Spec) == 1 {
		specMap, err := parseSpecBlock(cfg.Spec[0], hv)
		if err != nil {
			return nil, nil, fmt.Errorf("error in spec: %w", err)
		}
		pipeline[specKey] = specMap
	}

	if len(cfg.Includes) > 0 {
		includes, err := parseIncludeBlocks(cfg.Includes, hv)
		if err != nil {
			return nil, nil, err
		}

		pipeline["include"] = includes
	}

	for _, t := range cfg.Templates {
		templateMap, err := parseTemplateBlock(t, hv)
		if err != nil {
			return nil, nil, fmt.Errorf("error in template '%s': %w", t.ID, err)
		}

		key, err := blockKey(t.ID, t.Key, hv)
		if err != nil {
			return nil, nil, fmt.Errorf("error in template '%s': %w", t.ID, err)
		}

		if _, taken := jobs["."+key]; taken {
			return nil, nil, fmt.Errorf("duplicate template name '%s'", key)
		}

		keys["."+t.ID] = "." + key
		jobs["."+key] = templateMap
	}

	remapJobRefs(jobs, keys)

	if err := validatePipeline(pipeline, jobs); err != nil {
		return nil, nil, err
	}

	for name, job := range jobs {
		// The reserved keys are already in the map. A job named after one of
		// them used to overwrite it, so "stages" as a job name took the stage
		// list out of the file and the command still exited 0.
		//
		// The name is refused whether or not the pipeline happens to use that
		// keyword: GitLab reads "image:" at the top level as the keyword, not
		// as a job, so a pipeline whose only job was called "image" parsed to
		// YAML that unparse then read back as no jobs at all.
		if _, taken := pipeline[name]; taken || isReservedTopLevelKey(name) {
			return nil, nil, fmt.Errorf("%w: '%s'", errJobNamedAfterKeyword, name)
		}

		pipeline[name] = job
	}

	return pipeline, pipelineComments(body, keys, variableNames, variables, hv), nil
}

// pipelineComments reads the comments of the whole configuration, keyed by the
// YAML keys the blocks are written under rather than by their HCL labels. The
// two differ when a block carries an "id" attribute.
func pipelineComments(body hcl.Body, keys, variableNames map[string]string, variables map[string]any, hv *hclparser.HCLVars) *comments {
	out := readTopLevelComments(body, hv)

	if out == nil {
		return nil
	}

	relabel(out.child("variables"), variableNames)
	liftCollapsedVariables(out.child("variables"), variables)

	// A job and a template are written at the top level under their own key
	// rather than nested under "job" or "template", which is where the reader
	// left them.
	for _, blockType := range [...]string{"job", "template"} {
		nested := out.child(blockType)

		if nested == nil {
			continue
		}

		delete(out.children, blockType)

		for label, child := range nested.children {
			// A template is written with a leading dot, which is the label
			// the remap keyed it under.
			if blockType == "template" {
				label = "." + label
			}

			key, renamed := keys[label]

			if !renamed {
				key = label
			}

			out.set(key, child)
		}
	}

	return out
}

// liftCollapsedVariables moves the comments written inside a variable block up
// onto the key that block collapses to.
//
// A variable carrying nothing beyond its value is written out as a plain
// scalar, so the "name" and "value" attributes its comments were written on
// are not keys in the output and there is nothing left for them to sit on. One
// carrying more keeps both as real keys, and its comments stay where they are.
func liftCollapsedVariables(mapping *comments, variables map[string]any) {
	if mapping == nil {
		return
	}

	for key, child := range mapping.children {
		if _, expanded := variables[key].(map[string]any); expanded {
			continue
		}

		mapping.setOwn(key, nodeComment{
			head: firstNonEmpty(child.above(), child.at("name").head),
			line: child.at("value").line,
		})

		delete(mapping.children, key)
	}
}

// relabel rekeys a mapping's children from the block labels the reader used to
// the keys those blocks are written out under.
//
// A variable block is written under the value of its "name" attribute, and a
// job or template under its "id" when it carries one, so the label the comments
// arrived under is not where the emitter looks for them.
func relabel(mapping *comments, keys map[string]string) {
	if mapping == nil {
		return
	}

	// The renames are collected before any of them is applied: inserting into
	// a map while ranging over it leaves whether the new key is visited
	// undefined, so a rename whose target is another block's label was applied
	// twice on some runs and once on others, moving a comment onto the wrong
	// key. Same input, different output.
	type rename struct {
		from string
		to   string
		node *comments
	}

	renames := make([]rename, 0, len(mapping.children))

	for label, child := range mapping.children {
		key, renamed := keys[label]

		if !renamed || key == label {
			continue
		}

		renames = append(renames, rename{from: label, to: key, node: child})
	}

	for _, r := range renames {
		delete(mapping.children, r.from)
	}

	for _, r := range renames {
		mapping.children[r.to] = r.node
	}
}

// specKey is where a parsed "spec" header is kept until the YAML is written,
// at which point it becomes a document of its own ahead of the rest.
const specKey = "spec"

func parseSpecBlock(block hclSpecBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	for _, attr := range [...]struct {
		name string
		expr hcl.Expression
	}{
		{"inputs", block.Inputs},
		{"include", block.Include},
		{"component", block.Component},
		{"description", block.Description},
	} {
		if err := setOptionalAttr(out, attr.name, attr.expr, hv); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func parseVariableBlocks(blocks []hclVariableBlock, names map[string]string, hv *hclparser.HCLVars) (map[string]any, error) {
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

		// A variable carrying anything beyond its value is written back as
		// an object; a bare one stays a plain scalar.
		expanded := map[string]any{"value": value}

		for _, attr := range [...]struct {
			name string
			expr hcl.Expression
		}{
			{"description", b.Description},
			{"options", b.Options},
			{"expand", b.Expand},
		} {
			if err := setOptionalAttr(expanded, attr.name, attr.expr, hv); err != nil {
				return nil, fmt.Errorf("error in variable '%s': %w", b.ID, err)
			}
		}

		names[b.ID] = name

		if len(expanded) == 1 {
			result[name] = value

			continue
		}

		result[name] = expanded
	}

	return result, nil
}

func parseWorkflowBlock(block hclWorkflowBlock, hv *hclparser.HCLVars) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalAttr(out, "name", block.Name, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "auto_cancel", block.AutoCancel, hv); err != nil {
		return nil, err
	}

	rules, err := parseRuleBlocks(block.Rules, hv)
	if err != nil {
		return nil, err
	}

	if len(rules) > 0 {
		out["rules"] = rules
	} else if err := setEmptyCollection(out, "rules", block.EmptyRules, hv); err != nil {
		return nil, err
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

	if err := setOptionalAttr(out, "id_tokens", block.IDTokens, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "hooks", block.Hooks, hv); err != nil {
		return nil, err
	}

	if len(block.Artifacts) > 1 {
		return nil, errors.New("default can include at most one artifacts block")
	}

	if len(block.Artifacts) == 1 {
		artifacts, err := parseArtifactsBlock(block.Artifacts[0], hv)
		if err != nil {
			return nil, err
		}

		out["artifacts"] = artifacts
	} else if isExplicitNullExpr(block.EmptyArtifacts) {
		out["artifacts"] = nil
	}

	if cache, err := parseCacheBlocks(block.Cache, hv); err != nil {
		return nil, err
	} else if cache != nil {
		out["cache"] = cache
	} else if err := setEmptyCollection(out, "cache", block.EmptyCache, hv); err != nil {
		return nil, err
	}

	services, err := parseServiceBlocks(block.Services, hv)
	if err != nil {
		return nil, err
	}

	if len(services) > 0 {
		out["services"] = services
	} else if err := setEmptyCollection(out, "services", block.EmptyServices, hv); err != nil {
		return nil, err
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

	// The rest of the job keywords are plain expressions, kept in one list
	// so adding a keyword is a single line.
	for _, attr := range [...]struct {
		name string
		expr hcl.Expression
	}{
		{"dependencies", block.Dependencies},
		{"start_in", block.StartIn},
		{"identity", block.Identity},
		{"manual_confirmation", block.ManualConfirmation},
		{"inherit", block.Inherit},
		{"secrets", block.Secrets},
		{"id_tokens", block.IDTokens},
		{"hooks", block.Hooks},
		{"pages", block.Pages},
		{"run", block.Run},
		{"dast_configuration", block.DastConfiguration},
		{"inputs", block.Inputs},
		{"publish", block.Publish},
	} {
		if err := setOptionalAttr(out, attr.name, attr.expr, hv); err != nil {
			return nil, err
		}
	}

	// "only" and "except" may be written as null to clear what a job would
	// otherwise inherit, which is not the same as leaving them out.
	for _, attr := range [...]struct {
		name string
		expr hcl.Expression
	}{
		{"only", block.Only},
		{"except", block.Except},
	} {
		if err := setNullableAttr(out, attr.name, attr.expr, hv); err != nil {
			return nil, err
		}
	}

	if refs, err := parseReferenceList(block.DependsOn, "job"); err != nil {
		return nil, fmt.Errorf("depends_on: %w", err)
	} else if len(refs) > 0 {
		arr := make([]any, 0, len(refs))

		for _, ref := range refs {
			arr = append(arr, ref)
		}
		out["needs"] = arr
	} else if err := setEmptyCollection(out, "needs", block.DependsOn, hv); err != nil {
		return nil, err
	}

	if needs, err := parseNeedBlocks(block.Needs, hv); err != nil {
		return nil, err
	} else if len(needs) > 0 {
		out["needs"] = append(toAnyList(out["needs"]), needs...)
	}

	if refs, err := parseExtendsReferenceList(block.Extends); err != nil {
		return nil, fmt.Errorf("extends: %w", err)
	} else if len(refs) > 0 {
		arr := make([]any, 0, len(refs))

		for _, ref := range refs {
			arr = append(arr, ref)
		}
		out["extends"] = arr
	} else if isEmptyListExpr(block.Extends) {
		// An empty "extends" clears what a job would otherwise inherit, so it
		// is not the same as leaving the keyword out.
		out["extends"] = []any{}
	}

	rules, err := parseRuleBlocks(block.Rules, hv)
	if err != nil {
		return nil, err
	}

	if len(rules) > 0 {
		out["rules"] = rules
	} else if err := setEmptyCollection(out, "rules", block.EmptyRules, hv); err != nil {
		return nil, err
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
	} else if isExplicitNullExpr(block.EmptyArtifacts) {
		out["artifacts"] = nil
	}

	if cache, err := parseCacheBlocks(block.Cache, hv); err != nil {
		return nil, err
	} else if cache != nil {
		out["cache"] = cache
	} else if err := setEmptyCollection(out, "cache", block.EmptyCache, hv); err != nil {
		return nil, err
	}

	services, err := parseServiceBlocks(block.Services, hv)
	if err != nil {
		return nil, err
	}

	if len(services) > 0 {
		out["services"] = services
	} else if err := setEmptyCollection(out, "services", block.EmptyServices, hv); err != nil {
		return nil, err
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

		Dependencies:       block.Dependencies,
		StartIn:            block.StartIn,
		Identity:           block.Identity,
		ManualConfirmation: block.ManualConfirmation,
		Inherit:            block.Inherit,
		Secrets:            block.Secrets,
		IDTokens:           block.IDTokens,
		Hooks:              block.Hooks,
		Pages:              block.Pages,
		Run:                block.Run,
		DastConfiguration:  block.DastConfiguration,
		Inputs:             block.Inputs,
		Publish:            block.Publish,
		Only:               block.Only,
		Except:             block.Except,

		EmptyRules:     block.EmptyRules,
		EmptyCache:     block.EmptyCache,
		EmptyServices:  block.EmptyServices,
		EmptyArtifacts: block.EmptyArtifacts,

		Needs:     block.Needs,
		Rules:     block.Rules,
		Artifacts: block.Artifacts,
		Cache:     block.Cache,
		Services:  block.Services,
	}

	return parseJobBlock(job, hv)
}

// toAnyList returns a value already known to hold a list, or an empty one.
func toAnyList(value any) []any {
	list, _ := value.([]any)

	return list
}

// parseNeedBlocks turns "need" blocks back into the object form of a YAML
// "needs" entry. Its "job" is a job reference, like "depends_on", so the two
// forms name jobs the same way.
func parseNeedBlocks(blocks []hclNeedBlock, hv *hclparser.HCLVars) ([]any, error) {
	out := make([]any, 0, len(blocks))

	for _, block := range blocks {
		need := make(map[string]any)

		if refs, err := parseReferenceList(block.Job, "job"); err != nil {
			return nil, fmt.Errorf("need: job: %w", err)
		} else if len(refs) == 1 {
			need["job"] = refs[0]
		}

		for _, attr := range [...]struct {
			name string
			expr hcl.Expression
		}{
			{"artifacts", block.Artifacts},
			{"optional", block.Optional},
			{"parallel", block.Parallel},
			{"project", block.Project},
			{"ref", block.Ref},
			{"pipeline", block.Pipeline},
		} {
			if err := setOptionalAttr(need, attr.name, attr.expr, hv); err != nil {
				return nil, fmt.Errorf("need: %w", err)
			}
		}

		out = append(out, need)
	}

	return out, nil
}

func parseRuleBlocks(blocks []hclRuleBlock, hv *hclparser.HCLVars) ([]any, error) {
	out := make([]any, 0, len(blocks))

	for _, block := range blocks {
		rule := make(map[string]any)

		for _, attr := range [...]struct {
			name string
			expr hcl.Expression
		}{
			{"if", block.If},
			{"when", block.When},
			{"allow_failure", block.AllowFailure},
			{"changes", block.Changes},
			{"exists", block.Exists},
			{"variables", block.Variables},
			{"needs", block.Needs},
			{"start_in", block.StartIn},
			{"interruptible", block.Interruptible},
			{"auto_cancel", block.AutoCancel},
		} {
			if err := setOptionalAttr(rule, attr.name, attr.expr, hv); err != nil {
				return nil, err
			}
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

	if err := setOptionalAttr(out, "expose_as", block.ExposeAs, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "public", block.Public, hv); err != nil {
		return nil, err
	}

	if err := setOptionalAttr(out, "access", block.Access, hv); err != nil {
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

// parseCacheBlocks writes one cache block back as an object and several as a
// list, which is how GitLab spells multiple caches for one job.
func parseCacheBlocks(blocks []hclCacheBlock, hv *hclparser.HCLVars) (any, error) {
	if len(blocks) == 0 {
		return nil, nil
	}

	if len(blocks) == 1 {
		return parseCacheBlock(blocks[0], hv)
	}

	out := make([]any, 0, len(blocks))

	for _, block := range blocks {
		cache, err := parseCacheBlock(block, hv)
		if err != nil {
			return nil, err
		}

		out = append(out, cache)
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

	if err := setOptionalAttr(out, "unprotect", block.Unprotect, hv); err != nil {
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

		if err := setOptionalAttr(service, "docker", block.Docker, hv); err != nil {
			return nil, err
		}

		if err := setOptionalAttr(service, "kubernetes", block.Kubernetes, hv); err != nil {
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

		if err := setOptionalAttr(include, "rules", block.Rules, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		if err := setOptionalAttr(include, "integrity", block.Integrity, hv); err != nil {
			return nil, fmt.Errorf("error in include block: %w", err)
		}

		includes = append(includes, include)
	}

	if len(includes) == 1 {
		return includes[0], nil
	}

	return includes, nil
}

// setEmptyCollection records a keyword written as an explicit empty list, or
// as null. GitLab reads either as "override whatever this would inherit",
// which is not the same as leaving the keyword out, but a block has no empty
// spelling — so the schema carries an attribute alongside the block to hold
// it. The two spellings mean the same thing and both are kept as written.
func setEmptyCollection(out map[string]any, key string, expr hcl.Expression, hv *hclparser.HCLVars) error {
	if isExplicitNullExpr(expr) {
		out[key] = nil

		return nil
	}

	value, err := parseAttr(expr, hv)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}

	list, isList := value.([]any)

	if !isList {
		if value == nil {
			return nil
		}

		return fmt.Errorf("%s must be written as blocks unless it is an empty list", key)
	}

	if len(list) > 0 {
		return fmt.Errorf("%s must be written as blocks unless it is an empty list", key)
	}

	out[key] = []any{}

	return nil
}

// isEmptyListExpr reports whether an attribute was written as an empty list,
// rather than left out.
func isEmptyListExpr(expr hcl.Expression) bool {
	tuple, isTuple := expr.(*hclsyntax.TupleConsExpr)

	return isTuple && len(tuple.Exprs) == 0
}

// isExplicitNullExpr reports whether an attribute was written as "null",
// rather than left out. An absent optional attribute decodes to a placeholder
// expression that is not part of the file's syntax tree, which is what tells
// the two apart.
func isExplicitNullExpr(expr hcl.Expression) bool {
	if expr == nil {
		return false
	}

	literal, isLiteral := expr.(*hclsyntax.LiteralValueExpr)

	if !isLiteral {
		return false
	}

	return literal.Val.IsNull()
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

// setNullableAttr is setOptionalAttr for the few keywords GitLab lets you
// write as null to clear an inherited value. Everywhere else a null is not a
// legal spelling, so it stays dropped rather than being written back out.
func setNullableAttr(out map[string]any, key string, expr hcl.Expression, hv *hclparser.HCLVars) error {
	if isExplicitNullExpr(expr) {
		out[key] = nil

		return nil
	}

	return setOptionalAttr(out, key, expr, hv)
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

// blockKey returns the YAML key a job or template block is written under. An
// "id" attribute carries the original key when the label had to be sanitized to
// be referenceable, e.g. "build-app" becomes the label "build_app".
func blockKey(label string, expr hcl.Expression, hv *hclparser.HCLVars) (string, error) {
	value, err := parseAttr(expr, hv)
	if err != nil {
		return "", err
	}

	if value == nil {
		return label, nil
	}

	key, ok := value.(string)

	if !ok || key == "" {
		return "", errBlockIDNotString
	}

	return key, nil
}

// remapJobRefs rewrites "needs" and "extends" entries from block labels to the
// keys the jobs and templates are written under. They differ when a block
// carries an "id" attribute.
func remapJobRefs(jobs map[string]any, keys map[string]string) {
	for _, raw := range jobs {
		job, ok := raw.(map[string]any)

		if !ok {
			continue
		}

		for _, field := range [...]string{"needs", "extends"} {
			refs, ok := job[field].([]any)

			if !ok {
				continue
			}

			for i, ref := range refs {
				switch ref := ref.(type) {
				case string:
					if key, found := keys[ref]; found {
						refs[i] = key
					}

				case map[string]any:
					// The object form of a need holds the label under "job".
					// It used to be skipped, so a job whose name needed
					// sanitizing kept the sanitized name and the output did
					// not parse back.
					label, ok := ref["job"].(string)

					if !ok {
						continue
					}

					if key, found := keys[label]; found {
						ref["job"] = key
					}
				}
			}
		}
	}
}
