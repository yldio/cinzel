// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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

func parseHCLToWorkflows(body hcl.Body, sources map[string][]byte) ([]WorkflowYAMLFile, map[string]any, []ActionYAMLFile, error) {
	var cfg parseConfig
	diags := gohcl.DecodeBody(body, nil, &cfg)

	if diags.HasErrors() {
		return nil, nil, nil, cinzelerror.ProcessHCLDiags(diags)
	}

	hv := hclparser.NewHCLVars()
	hv.SetSources(sources)

	if err := cfg.Variables.Parse(hv); err != nil {
		return nil, nil, nil, err
	}

	// Ahead of the decode below, which finds the same collision but cannot
	// say where either block is.
	if err := checkDuplicateStepLabels(body); err != nil {
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
	jobComments := labelledBlockComments(hv, labelledBlocks(body, "job"))

	for _, j := range cfg.Jobs {
		// Two blocks with the same label used to overwrite one another in the
		// map, leaving the last one and no word about the rest.
		if _, taken := parsedJobs[j.ID]; taken {
			return nil, nil, nil, fmt.Errorf("%w: '%s' is declared twice", errDuplicateJobLabel, j.ID)
		}

		job, err := parseJobConfig(j, hv)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error in job '%s': %w", j.ID, err)
		}

		if len(job.StepRefs) > 0 {
			steps := make([]any, 0, len(job.StepRefs))
			emitted := make(map[string]struct{}, len(job.StepRefs))

			// Two step blocks can reach the same "id" through their own "id"
			// attributes, which GitHub refuses: a step id has to be unique
			// within its job. That went out as a written file and exit 0, and
			// only actionlint or GitHub itself said otherwise.
			takenIDs := make(map[string]string, len(job.StepRefs))

			for _, stepID := range job.StepRefs {
				stepVal, exists := stepMap[stepID]

				if !exists {
					return nil, nil, nil, fmt.Errorf("error in job '%s': cannot find step '%s'", j.ID, stepID)
				}

				// A job may run the same step more than once. GitHub requires a
				// step id to be unique within its job, so only the first
				// occurrence carries one.
				if _, repeat := emitted[stepID]; repeat {
					stepVal = stepValueWithoutID(stepVal)
				} else if id, ok := emittedStepID(stepVal); ok {
					if other, taken := takenIDs[id]; taken {
						return nil, nil, nil, fmt.Errorf("error in job '%s': %w: '%s' and '%s' both write '%s'",
							j.ID, errDuplicateStepID, other, stepID, id)
					}

					takenIDs[id] = stepID
				}

				if err := checkStepNotEmpty(stepVal, stepID); err != nil {
					return nil, nil, nil, fmt.Errorf("error in job '%s': %w", j.ID, err)
				}

				emitted[stepID] = struct{}{}

				steps = append(steps, stepVal)
			}

			job.Body["steps"] = steps
		}

		parsedJobs[j.ID] = job
	}

	if err := validateParsedJobs(parsedJobs); err != nil {
		return nil, nil, nil, err
	}

	remapJobNeeds(parsedJobs)

	parsedWorkflows := make([]WorkflowYAMLFile, 0, len(cfg.Workflows))
	takenFilenames := make(map[string]string, len(cfg.Workflows))
	workflowComments := labelledBlockComments(hv, labelledBlocks(body, "workflow"))

	for _, wf := range cfg.Workflows {
		workflow, err := parseWorkflowConfig(wf, hv)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w", wf.ID, err)
		}

		if workflow.Filename == "" {
			return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w", wf.ID, cinzelerror.ErrWorkflowFilenameRequired)
		}

		if err := checkFilenameStaysInside(workflow.Filename); err != nil {
			return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w", wf.ID, err)
		}

		if err := claimFilename(takenFilenames, workflow.Filename, wf.ID); err != nil {
			return nil, nil, nil, err
		}

		if len(workflow.JobRefs) > 0 {
			jobs := make(map[string]any)

			jobOrder := make([]string, 0, len(workflow.JobRefs))

			// Two jobs can reach the same YAML key through their "id"
			// attributes, which used to leave one of them out of the file
			// silently. The check is per workflow: each is its own file, so
			// the same key in another one is not a collision.
			takenKeys := make(map[string]string, len(workflow.JobRefs))

			for _, jobID := range workflow.JobRefs {
				jobContent, exists := parsedJobs[jobID]

				if !exists {
					return nil, nil, nil, fmt.Errorf("error in workflow '%s': cannot find job '%s'", wf.ID, jobID)
				}

				if other, taken := takenKeys[jobContent.Key]; taken {
					return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w: '%s' and '%s' both write '%s'",
						wf.ID, errDuplicateJobKey, other, jobID, jobContent.Key)
				}
				takenKeys[jobContent.Key] = jobID

				// The comments were written on the block, so they are keyed by
				// the block label rather than by the YAML key the job lands
				// under, which may have been set by an "id" attribute.
				jobs[jobContent.Key] = withComments(jobContent.Body, jobComments[jobID])

				jobOrder = append(jobOrder, jobContent.Key)
			}

			workflow.Body["jobs"] = jobs
			workflow.JobRefs = jobOrder
		}

		// Validation runs here rather than above, because every check that
		// walks jobs and steps — the expression syntax check among them — saw
		// an empty workflow while "jobs" was still unassigned, so they were
		// dead on the parse path.
		if err := validateParsedWorkflow(workflow); err != nil {
			return nil, nil, nil, fmt.Errorf("error in workflow '%s': %w", wf.ID, err)
		}

		parsedWorkflows = append(parsedWorkflows, WorkflowYAMLFile{
			Filename:    workflow.Filename,
			Content:     workflow.Body,
			JobOrder:    workflow.JobRefs,
			FootComment: workflowComments[wf.ID].foot,
		})
	}

	parsedActions, err := parseHCLActions(cfg.Actions, body, hv, stepMap)
	if err != nil {
		return nil, nil, nil, err
	}

	return parsedWorkflows, stepMap, parsedActions, nil
}

// remapJobNeeds rewrites "needs" entries from block labels to the YAML keys the
// jobs are written under. They differ when a job carries an "id" attribute.
func remapJobNeeds(jobs map[string]ghjob.Parsed) {
	keys := make(map[string]string, len(jobs))

	for label, job := range jobs {
		keys[label] = job.Key
	}

	for _, job := range jobs {
		needs, ok := job.Body["needs"].([]any)

		if !ok {
			continue
		}

		for i, dep := range needs {
			label, ok := dep.(string)

			if !ok {
				continue
			}

			if key, found := keys[label]; found {
				needs[i] = key
			}
		}
	}
}

// parseJobConfig converts one job block. Step references are returned as a
// field rather than smuggled through the body map.
func parseJobConfig(cfg hclJobBlock, hv *hclparser.HCLVars) (ghjob.Parsed, error) {
	job := ghjob.Parsed{ID: cfg.ID, Key: cfg.ID, Body: map[string]any{}}
	out := job.Body

	// An "id" attribute carries the original YAML key when it was sanitized
	// to make the block referenceable, e.g. "build-and-test". The label stays
	// the reference name.
	key, err := parseAttr(cfg.Key, hv)
	if err != nil {
		return ghjob.Parsed{}, err
	}

	if key != nil {
		str, ok := key.(string)

		if !ok || str == "" {
			return ghjob.Parsed{}, errJobIDNotString
		}

		job.Key = str
	}

	if err := setOptionalYAMLAttr(out, "name", cfg.Name, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "if", cfg.If, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "uses", cfg.Uses, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	stepRefs, err := parseReferenceList(cfg.Steps, "step")
	if err != nil {
		return ghjob.Parsed{}, err
	}

	job.StepRefs = stepRefs

	if refs, err := parseReferenceList(cfg.DependsOn, "job"); err != nil {
		return ghjob.Parsed{}, err
	} else if len(refs) > 0 {
		deps := make([]any, 0, len(refs))

		for _, ref := range refs {
			deps = append(deps, ref)
		}

		out["needs"] = deps
	}

	if err := setOptionalYAMLAttr(out, "secrets", cfg.Secrets, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "continue-on-error", cfg.ContinueOnError, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "timeout-minutes", cfg.TimeoutMinutes, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	for _, usesBlock := range cfg.UsesBlocks {
		usesValue, err := parseUsesBlockFromConfig(usesBlock, hv)
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "uses", "uses"); err != nil {
			return ghjob.Parsed{}, err
		}

		out["uses"] = usesValue
	}

	for _, block := range cfg.WithBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {
			return ghjob.Parsed{}, err
		}

		withMap := getOrCreateMap(out, "with")

		if err := claimBlockKey(withMap, "with", key); err != nil {
			return ghjob.Parsed{}, err
		}

		withMap[key] = withComments(value, namedBlockComments(block, hv))
	}

	for _, block := range cfg.EnvBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {
			return ghjob.Parsed{}, err
		}

		envMap := getOrCreateMap(out, "env")

		if err := claimBlockKey(envMap, "env", key); err != nil {
			return ghjob.Parsed{}, err
		}

		envMap[key] = withComments(value, namedBlockComments(block, hv))
	}

	for _, block := range cfg.OutputBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {
			return ghjob.Parsed{}, err
		}

		outputsMap := getOrCreateMap(out, "outputs")

		if err := claimBlockKey(outputsMap, "output", key); err != nil {
			return ghjob.Parsed{}, err
		}

		outputsMap[key] = withComments(value, namedBlockComments(block, hv))
	}

	for _, block := range cfg.SecretBlocks {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {
			return ghjob.Parsed{}, err
		}

		secretsMap := getOrCreateMap(out, "secrets")

		if err := claimBlockKey(secretsMap, "secret", key); err != nil {
			return ghjob.Parsed{}, err
		}

		secretsMap[key] = withComments(value, namedBlockComments(block, hv))
	}

	for _, block := range cfg.ServiceBlocks {
		serviceVal, err := parseBodyMap(block.Body, hv, "service")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := checkBlockNotEmpty("service", serviceVal); err != nil {
			return ghjob.Parsed{}, err
		}

		servicesMap := getOrCreateMap(out, "services")

		if err := claimBlockKey(servicesMap, "service", block.ID); err != nil {
			return ghjob.Parsed{}, err
		}

		servicesMap[block.ID] = withComments(serviceVal, blockComments(block.Body, hv))
	}

	for _, block := range cfg.RunsOnBlocks {
		runsOnValue, err := parseBodyMap(block.Body, hv, "runs_on")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		comments := blockComments(block.Body, hv)

		if err := checkBlockNotEmpty("runs_on", runsOnValue); err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "runs_on", "runs-on"); err != nil {
			return ghjob.Parsed{}, err
		}

		if runners, ok := runsOnValue["runners"]; ok && len(runsOnValue) == 1 {
			out["runs-on"] = withComments(runners, comments)
		} else {
			out["runs-on"] = withComments(runsOnValue, comments)
		}
	}

	for _, block := range cfg.StrategyBlocks {
		strategyValue, err := parseBodyMap(block.Body, hv, "strategy")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := checkBlockNotEmpty("strategy", strategyValue); err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "strategy", "strategy"); err != nil {
			return ghjob.Parsed{}, err
		}

		out["strategy"] = withComments(strategyValue, blockComments(block.Body, hv))
	}

	if err := setOptionalYAMLAttr(out, "permissions", cfg.PermAttr, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "concurrency", cfg.ConcurrencyAttr, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "container", cfg.ContainerAttr, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "environment", cfg.EnvironmentAttr, hv); err != nil {
		return ghjob.Parsed{}, err
	}

	for _, block := range cfg.Permissions {
		child, err := parseBodyMap(block.Body, hv, "permissions")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := checkBlockNotEmpty("permissions", child); err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "permissions", "permissions"); err != nil {
			return ghjob.Parsed{}, err
		}

		out["permissions"] = withComments(child, blockComments(block.Body, hv))
	}

	for _, block := range cfg.Defaults {
		child, err := parseBodyMap(block.Body, hv, "defaults")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := checkBlockNotEmpty("defaults", child); err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "defaults", "defaults"); err != nil {
			return ghjob.Parsed{}, err
		}

		out["defaults"] = withComments(child, blockComments(block.Body, hv))
	}

	for _, block := range cfg.Concurrency {
		child, err := parseBodyMap(block.Body, hv, "concurrency")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := checkBlockNotEmpty("concurrency", child); err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "concurrency", "concurrency"); err != nil {
			return ghjob.Parsed{}, err
		}

		out["concurrency"] = withComments(child, blockComments(block.Body, hv))
	}

	for _, block := range cfg.Container {
		child, err := parseBodyMap(block.Body, hv, "container")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := checkBlockNotEmpty("container", child); err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "container", "container"); err != nil {
			return ghjob.Parsed{}, err
		}

		out["container"] = withComments(child, blockComments(block.Body, hv))
	}

	for _, block := range cfg.Environment {
		child, err := parseBodyMap(block.Body, hv, "environment")
		if err != nil {
			return ghjob.Parsed{}, err
		}

		if err := checkBlockNotEmpty("environment", child); err != nil {
			return ghjob.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "environment", "environment"); err != nil {
			return ghjob.Parsed{}, err
		}

		out["environment"] = withComments(child, blockComments(block.Body, hv))
	}

	return job, nil
}

// parseWorkflowConfig converts one workflow block. The filename and the job
// references are returned as fields rather than smuggled through the body map.
func parseWorkflowConfig(cfg hclWorkflowBlock, hv *hclparser.HCLVars) (ghworkflow.Parsed, error) {
	workflow := ghworkflow.Parsed{ID: cfg.ID, Body: map[string]any{}}
	out := workflow.Body

	filename, err := parseAttr(cfg.Filename, hv)
	if err != nil {
		return ghworkflow.Parsed{}, err
	}

	workflow.Filename, _ = filename.(string)

	if err := setOptionalYAMLAttr(out, "name", cfg.Name, hv); err != nil {
		return ghworkflow.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "run-name", cfg.RunName, hv); err != nil {
		return ghworkflow.Parsed{}, err
	}

	jobRefs, err := parseReferenceList(cfg.Jobs, "job")
	if err != nil {
		return ghworkflow.Parsed{}, err
	}

	workflow.JobRefs = jobRefs

	if err := setOptionalYAMLAttr(out, "permissions", cfg.Permissions, hv); err != nil {
		return ghworkflow.Parsed{}, err
	}

	if err := setOptionalYAMLAttr(out, "concurrency", cfg.Concurrency, hv); err != nil {
		return ghworkflow.Parsed{}, err
	}

	for _, on := range cfg.On {
		eventValue, err := parseBodyMap(on.Body, hv, "on")
		if err != nil {
			return ghworkflow.Parsed{}, err
		}

		eventName := on.ID
		eventValue = ghworkflow.NormalizeOnEvent(eventName, eventValue)
		comments := blockComments(on.Body, hv)

		onMap := getOrCreateMap(out, "on")

		if err := claimBlockKey(onMap, "on", eventName); err != nil {
			return ghworkflow.Parsed{}, err
		}

		if eventName == "schedule" {
			onMap[eventName] = withComments(ghworkflow.DenormalizeScheduleEvent(eventValue), comments)
		} else if len(eventValue) == 0 {
			onMap[eventName] = withComments(map[string]any{}, comments)
		} else {
			onMap[eventName] = withComments(eventValue, comments)
		}
	}

	for _, block := range cfg.Env {
		key, value, err := parseNamedConfig(block, hv)
		if err != nil {
			return ghworkflow.Parsed{}, err
		}

		envMap := getOrCreateMap(out, "env")

		if err := claimBlockKey(envMap, "env", key); err != nil {
			return ghworkflow.Parsed{}, err
		}

		envMap[key] = withComments(value, namedBlockComments(block, hv))
	}

	for _, block := range cfg.PermBlocks {
		child, err := parseBodyMap(block.Body, hv, "permissions")
		if err != nil {
			return ghworkflow.Parsed{}, err
		}

		if err := checkBlockNotEmpty("permissions", child); err != nil {
			return ghworkflow.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "permissions", "permissions"); err != nil {
			return ghworkflow.Parsed{}, err
		}

		out["permissions"] = withComments(child, blockComments(block.Body, hv))
	}

	if _, ok := out["permissions"]; !ok {
		out["permissions"] = map[string]any{}
	}

	for _, block := range cfg.Defaults {
		child, err := parseBodyMap(block.Body, hv, "defaults")
		if err != nil {
			return ghworkflow.Parsed{}, err
		}

		if err := checkBlockNotEmpty("defaults", child); err != nil {
			return ghworkflow.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "defaults", "defaults"); err != nil {
			return ghworkflow.Parsed{}, err
		}

		out["defaults"] = withComments(child, blockComments(block.Body, hv))
	}

	for _, block := range cfg.ConcBlocks {
		child, err := parseBodyMap(block.Body, hv, "concurrency")
		if err != nil {
			return ghworkflow.Parsed{}, err
		}

		if err := checkBlockNotEmpty("concurrency", child); err != nil {
			return ghworkflow.Parsed{}, err
		}

		if err := claimSingletonBlock(out, "concurrency", "concurrency"); err != nil {
			return ghworkflow.Parsed{}, err
		}

		out["concurrency"] = withComments(child, blockComments(block.Body, hv))
	}

	return workflow, nil
}

func setOptionalYAMLAttr(out map[string]any, yamlKey string, expr hcl.Expression, hv *hclparser.HCLVars) error {
	val, err := parseAttr(expr, hv)
	if err != nil {
		return err
	}

	if val == nil {
		return nil
	}

	// The expression range starts at the value and ends where it ends, which
	// is the same line the attribute is written on either way, so a comment
	// above it or beside it is found from here without the attribute itself.
	out[yamlKey] = annotate(val, hv, expr.Range())

	return nil
}

// blockHeadComment returns the comment written above the block whose body this
// is.
//
// The typed decode hands back a body and not the block header, so the line to
// look above is the body's own start: that is the open brace, which shares a
// line with the block type in every block these providers write.
func blockHeadComment(body hcl.Body, hv *hclparser.HCLVars) string {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {
		return ""
	}

	return hv.HeadComment(sb.SrcRange)
}

// blockFootComment returns the comment written at the end of a block body with
// nothing after it. The body's end range is its closing brace, and the run
// above that brace is the foot: the same read as a head comment, taken from
// the other end of the block.
func blockFootComment(body hcl.Body, hv *hclparser.HCLVars) string {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {
		return ""
	}

	return hv.HeadComment(sb.EndRange)
}

// blockComment holds what was written above a block and at the end of its
// body. The two travel together because they are two halves of one block's
// comments, and a caller that has one has the block the other is read from.
type blockComment struct {
	head string
	foot string
	// trailing is the comment sharing a line with the value the block
	// resolves to. A block has no line of its own to carry one, so this is
	// only set where the value inside it was read.
	trailing string
}

// blockComments reads both of a block body's comments.
func blockComments(body hcl.Body, hv *hclparser.HCLVars) blockComment {
	return blockComment{
		head: blockHeadComment(body, hv),
		foot: blockFootComment(body, hv),
	}
}

// namedBlockComments reads every comment a name/value block carries.
//
// There are two places a run of comments sits above one of these: above the
// block, and above the value inside it. The block becomes a single YAML key,
// so both belong above that key, joined in the order they were written. Only
// the first used to be read, so a note written above the "value =" of a job's
// env block was dropped while the step path kept the same note. The trailing
// comment on the value is the block's too, for the same reason.
func namedBlockComments(cfg hclNamedBlock, hv *hclparser.HCLVars) blockComment {
	c := blockComments(cfg.Body, hv)

	if cfg.Value == nil {
		return c
	}

	r := cfg.Value.Range()

	return blockComment{
		head:     joinComments(c.head, hv.HeadComment(r)),
		trailing: hv.TrailingComment(r),
		foot:     c.foot,
	}
}

// joinComments runs two comment runs together, dropping an empty one. Each is
// already a run of whole lines, so one newline between them reads as one run.
func joinComments(above, below string) string {
	switch {
	case above == "":
		return below
	case below == "":
		return above
	default:
		return above + "\n" + below
	}
}

// withComments wraps val with the comments written on the block it came from,
// returning val unwrapped when there were none. The wrapper is what carries a
// comment to the emitter, and an unconditional one would put every value
// behind it for nothing.
func withComments(val any, c blockComment) any {
	if c.head == "" && c.foot == "" && c.trailing == "" {
		return val
	}

	return annotated{value: val, head: c.head, foot: c.foot, comment: c.trailing}
}

// annotate wraps val with whatever comments were written above or beside r,
// returning val unwrapped when there were none. The wrapper is what carries a
// comment to the emitter, and an unconditional one would put every value
// behind it for nothing.
func annotate(val any, hv *hclparser.HCLVars, r hcl.Range) any {
	trailing := hv.TrailingComment(r)
	head := hv.HeadComment(r)

	if trailing == "" && head == "" {
		return val
	}

	return annotated{value: val, comment: trailing, head: head}
}

func parseNamedConfig(cfg hclNamedBlock, hv *hclparser.HCLVars) (string, any, error) {
	if err := hclparser.RejectUnknown(cfg.Body); err != nil {
		return "", nil, err
	}

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
	val, _, err := list.Parse(hv)
	if err != nil {
		return "", err
	}

	return val.AsString(), nil
}

// yamlKeyIn spells an HCL key the way YAML wants it, except where the key is
// the author's own data rather than one of the provider's.
//
// A matrix axis is named by whoever wrote it, and referenced by that name in
// "${{ matrix.X }}", which cinzel copies through as written. Renaming the axis
// and not the reference left the reference resolving against an axis that no
// longer existed: actionlint reads it as `property "go_version" is not defined
// in object type {go-version: ...}`, and the run fails at the reference. The
// unparse side already writes an axis name back unchanged, so the rename was
// also the one thing a roundtrip could not undo.
func yamlKeyIn(scope, name string) string {
	if scope == "matrix" {
		return name
	}

	return naming.ToYAMLKey(name)
}

// childScope names the scope a nested block's body is read in.
//
// Everything written inside a matrix is named by whoever wrote the workflow,
// however deep it sits: a "variable" block may list its axes directly rather
// than through "name" and "value". The scope was replaced by the block's own
// type on the way down, so those names left the matrix scope and were renamed
// with it — "go_version" came out "go-version", which no
// "${{ matrix.go_version }}" reference names, while the same axis written as a
// matrix attribute kept its name. The matrix scope carries down instead.
func childScope(scope, blockType string) string {
	if scope == "matrix" {
		return "matrix"
	}

	return blockType
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

			out[yamlKeyIn(scope, name)] = annotate(val, hv, attr.SrcRange)
		}
	}

	for _, block := range sb.Blocks {
		// A block becomes a key in the YAML, so the comment above it belongs
		// above that key. Read once here rather than in each case below,
		// which differ in where the key lands and not in where its comment
		// was written.
		// The block is in hand rather than just its body, so the head comes
		// off its type range: the line the block opens on. The foot is read
		// from the body, which is where the closing brace is.
		comments := blockComment{
			head: hv.HeadComment(block.TypeRange),
			foot: blockFootComment(block.Body, hv),
		}

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

			if err := claimBlockKey(onMap, "on", eventName); err != nil {
				return nil, err
			}

			if eventName == "schedule" {
				onMap[eventName] = withComments(ghworkflow.DenormalizeScheduleEvent(eventValue), comments)
			} else if len(eventValue) == 0 {
				onMap[eventName] = withComments(map[string]any{}, comments)
			} else {
				onMap[eventName] = withComments(eventValue, comments)
			}
		case block.Type == "uses":
			usesValue, err := parseUsesBlock(block.Body, hv)
			if err != nil {
				return nil, err
			}

			if err := claimSingletonBlock(out, "uses", "uses"); err != nil {
				return nil, err
			}

			out["uses"] = withComments(usesValue, comments)
		case block.Type == "with":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {
				return nil, err
			}

			withMap := getOrCreateMap(out, "with")

			if err := claimBlockKey(withMap, "with", key); err != nil {
				return nil, err
			}

			withMap[key] = withComments(value, comments)
		case block.Type == "env":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {
				return nil, err
			}

			envMap := getOrCreateMap(out, "env")

			if err := claimBlockKey(envMap, "env", key); err != nil {
				return nil, err
			}

			envMap[key] = withComments(value, comments)
		case block.Type == "output" && scope == "job":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {
				return nil, err
			}

			outputsMap := getOrCreateMap(out, "outputs")

			if err := claimBlockKey(outputsMap, "output", key); err != nil {
				return nil, err
			}

			outputsMap[key] = withComments(value, comments)
		case block.Type == "secret" && scope == "job":
			key, value, err := parseNamedBlock(block.Body, hv)
			if err != nil {
				return nil, err
			}

			secretsMap := getOrCreateMap(out, "secrets")

			if err := claimBlockKey(secretsMap, "secret", key); err != nil {
				return nil, err
			}

			secretsMap[key] = withComments(value, comments)
		case block.Type == "service" && scope == "job":
			if len(block.Labels) != 1 {
				return nil, errors.New("service block must have exactly one label")
			}

			serviceVal, err := parseBodyMap(block.Body, hv, "service")
			if err != nil {
				return nil, err
			}

			if err := checkBlockNotEmpty("service", serviceVal); err != nil {
				return nil, err
			}

			servicesMap := getOrCreateMap(out, "services")

			if err := claimBlockKey(servicesMap, "service", block.Labels[0]); err != nil {
				return nil, err
			}

			servicesMap[block.Labels[0]] = withComments(serviceVal, comments)
		case block.Type == "runs_on" && scope == "job":
			runsOnValue, err := parseBodyMap(block.Body, hv, "runs_on")
			if err != nil {
				return nil, err
			}

			if err := checkBlockNotEmpty("runs_on", runsOnValue); err != nil {
				return nil, err
			}

			if err := claimSingletonBlock(out, "runs_on", "runs-on"); err != nil {
				return nil, err
			}

			if runners, ok := runsOnValue["runners"]; ok && len(runsOnValue) == 1 {
				out["runs-on"] = withComments(runners, comments)
			} else {
				out["runs-on"] = withComments(runsOnValue, comments)
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

			if err := checkBlockNotEmpty("matrix", normalized); err != nil {
				return nil, err
			}

			if err := claimSingletonBlock(out, "matrix", "matrix"); err != nil {
				return nil, err
			}

			out["matrix"] = withComments(normalized, comments)
		default:
			child, err := parseBodyMap(block.Body, hv, childScope(scope, block.Type))
			if err != nil {
				return nil, err
			}

			if err := checkBlockNotEmpty(block.Type, child); err != nil {
				return nil, err
			}

			if err := addGenericBlock(out, yamlKeyIn(scope, block.Type), block.Type, block.Labels, withComments(child, comments)); err != nil {
				return nil, err
			}
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
	val, _, err := list.Parse(hv)
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

// addGenericBlock files a nested block under key. A labelled one is keyed by
// its label, so two sharing a label are refused rather than written one over
// the other. An unlabelled one is appended, which is the shape the caller wants
// there.
func addGenericBlock(target map[string]any, key, blockType string, labels []string, value any) error {
	if len(labels) == 1 {
		mapping, ok := target[key].(map[string]any)

		if !ok || mapping == nil {
			mapping = map[string]any{}
		}

		if err := claimBlockKey(mapping, blockType, labels[0]); err != nil {
			return err
		}

		mapping[labels[0]] = value
		target[key] = mapping

		return nil
	}

	if existing, ok := target[key]; ok {
		switch casted := existing.(type) {
		case []any:
			target[key] = append(casted, value)
		default:
			target[key] = []any{casted, value}
		}

		return nil
	}

	target[key] = value

	return nil
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

// claimSingletonBlock refuses a second block of a type that writes one whole
// key, and a block written alongside the attribute spelling of the same key.
//
// Unlike a named block, one of these carries no key of its own: it simply
// assigns out[key], so a second one replaced the first and the run still
// exited 0. The permissions a workflow shipped with were then whichever of the
// two the decoder happened to hand back last.
func claimSingletonBlock(target map[string]any, blockType, key string) error {
	if _, taken := target[key]; taken {
		return fmt.Errorf("%w: '%s' is written more than once", errDuplicateBlock, blockType)
	}

	return nil
}

// claimBlockKey refuses a second block writing a key an earlier one already
// wrote.
//
// The key comes from whoever wrote the file — a "name" attribute, a block
// label — so two blocks are free to name the same one, and the map simply took
// the last. The rest went out neither written nor reported, so a workflow
// shipped with a value nobody in the file meant to set. A job and a step
// already refuse the same input, and so does the GitLab provider for its
// variable blocks.
func claimBlockKey(target map[string]any, blockType, key string) error {
	if _, taken := target[key]; taken {
		return fmt.Errorf("%w: two '%s' blocks both write '%s'", errDuplicateBlockKey, blockType, key)
	}

	return nil
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

// emittedStepID returns the "id" a converted step writes, and whether it writes
// one at all. A step carrying ignore_id has none.
//
// A step written under a comment arrives wrapped, so the map is reached through
// the wrapper rather than asserted straight off the value: asserting alone read
// every commented step as having no id, and let two of them write the same one.
func emittedStepID(stepVal any) (string, bool) {
	m, ok := stepMapValue(stepVal)
	if !ok {
		return "", false
	}

	id, ok := plain(m["id"]).(string)

	return id, ok && id != ""
}

// stepValueWithoutID copies a converted step with its "id" left out, for a
// repeat occurrence of a step within one job. A wrapped step keeps its wrapper,
// so the comments written above it survive the copy.
func stepValueWithoutID(stepVal any) any {
	if wrapper, ok := stepVal.(annotated); ok {
		wrapper.value = stepValueWithoutID(wrapper.value)

		return wrapper
	}

	m, ok := stepVal.(map[string]any)
	if !ok {
		return stepVal
	}

	out := make(map[string]any, len(m))

	for key, value := range m {
		if key != "id" {
			out[key] = value
		}
	}

	return out
}

// checkBlockNotEmpty refuses a block that converted to nothing at all.
//
// A block with an empty body becomes an empty map, which goes out as a bare
// "key:" with a null under it. GitHub rejects every one of these — actionlint
// reads "runs-on:" as `string should not be empty` and "defaults:" as
// `"defaults" section should have "run" section` — and cinzel's own unparse
// refuses the runs-on form, so a parse exited 0 on a file nothing downstream
// could read. "permissions" is the exception: an empty one is written
// explicitly as "{}", because an absent permissions field makes GitHub inherit
// the default token permissions rather than granting none.
func checkBlockNotEmpty(blockType string, value map[string]any) error {
	if blockType == "permissions" {
		return nil
	}

	if len(value) == 0 {
		return fmt.Errorf("%w: '%s' sets nothing", errEmptyBlock, blockType)
	}

	return nil
}

// checkStepNotEmpty refuses a step that converted to nothing at all.
//
// A step block whose every attribute is absent, or that carries only
// "ignore_id", leaves an empty map, which goes out as a bare "-" under
// "steps". GitHub rejects that file, and so does cinzel's own unparse, which
// reads the null node back and reports a step that must be an object. The
// check sits where the step is used rather than where it is declared: a step
// block nothing references is never emitted and is nobody's problem.
func checkStepNotEmpty(stepVal any, stepID string) error {
	m, ok := stepMapValue(stepVal)

	if ok && len(m) == 0 {
		return fmt.Errorf("%w: '%s' sets nothing", errEmptyStep, stepID)
	}

	return nil
}

// stepMapValue returns the map a converted step is, reaching through the
// annotated wrapper a commented step carries.
func stepMapValue(stepVal any) (map[string]any, bool) {
	if wrapper, ok := stepVal.(annotated); ok {
		return stepMapValue(wrapper.value)
	}

	m, ok := stepVal.(map[string]any)

	return m, ok
}

// annotateStep attaches the comments read off a step block to the converted
// step, wrapping each attribute that carried one and the step itself if a
// comment was written above the block.
func annotateStep(converted any, comments step.Comments) any {
	m, ok := converted.(map[string]any)

	if !ok {
		return converted
	}

	for key, value := range m {
		comment := comments.At(key)

		if comment.Empty() {
			continue
		}

		m[key] = annotated{value: value, comment: comment.Line, head: comment.Head}
	}

	for key, entries := range comments.Nested {
		nested, ok := m[key].(map[string]any)

		if !ok {
			continue
		}

		for name, comment := range entries {
			value, exists := nested[name]

			if !exists {
				continue
			}

			nested[name] = annotated{value: value, comment: comment.Line, head: comment.Head, foot: comment.Foot}
		}
	}

	return withComments(m, blockComment{head: comments.Head, foot: comments.Foot})
}

func stepsToMap(steps step.Steps) (map[string]any, error) {
	out := make(map[string]any, len(steps))

	for stepID, parsedStep := range steps {
		converted, err := yamlwriter.Convert(parsedStep)
		if err != nil {
			return nil, err
		}

		out[stepID] = annotateStep(converted, parsedStep.Comments)
	}

	return out, nil
}

// checkDuplicateStepLabels reports two step blocks sharing a label.
//
// A step label is the address `step.<label>` resolves, so a repeat leaves the
// reference ambiguous. step.Parse catches this too, but it works from decoded
// values and has no source position to report, which sent the author hunting
// through every file in the directory. Read here from the block headers, where
// the ranges survive the merge.
func checkDuplicateStepLabels(body hcl.Body) error {
	content, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "step", LabelNames: []string{"id"}}},
	})

	// Left to the full decode, which reports it with the detail this pass
	// deliberately does not collect.
	if diags.HasErrors() {
		return nil
	}

	seen := make(map[string]hcl.Range, len(content.Blocks))

	for _, block := range content.Blocks {
		label := block.Labels[0]

		if first, taken := seen[label]; taken {
			return fmt.Errorf("%w: '%s' is declared at %s and again at %s",
				errDuplicateStepLabel, label, first, block.DefRange)
		}

		seen[label] = block.DefRange
	}

	return nil
}
