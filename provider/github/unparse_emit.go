// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/hclcomment"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
	"github.com/zclconf/go-cty/cty"
)

type workflowJobEntry struct {
	Name string
	Body map[string]any
}

func buildWorkflowJobIndex(jobs map[string]any, order []string, usedRefs map[string]struct{}) ([]workflowJobEntry, []string, map[string]string, error) {
	jobNames := order
	if len(jobNames) == 0 {
		jobNames = sortedKeys(jobs)
	}
	entries := make([]workflowJobEntry, 0, len(jobs))
	jobRefs := make([]string, 0, len(jobs))
	jobIDMap := make(map[string]string, len(jobs))

	for _, jobName := range jobNames {
		raw, exists := jobs[jobName]

		if !exists {
			return nil, nil, nil, fmt.Errorf("job '%s' listed in order but not found in jobs map", jobName)
		}

		jobMap, ok := toStringAnyMap(raw)

		if !ok {
			return nil, nil, nil, fmt.Errorf("job '%s' must be an object", jobName)
		}

		entries = append(entries, workflowJobEntry{Name: jobName, Body: jobMap})

		jobID := sanitizeIdentifier(jobName)

		if jobID == "" {
			jobID = "job"
		}

		jobID = uniqueIdentifierInSet(jobID, usedRefs)
		usedRefs[jobID] = struct{}{}
		jobRefs = append(jobRefs, jobID)
		jobIDMap[jobName] = jobID
	}

	for jobName := range jobs {
		if _, covered := jobIDMap[jobName]; !covered {
			return nil, nil, nil, fmt.Errorf("job '%s' is defined but was not included in the job order", jobName)
		}
	}

	return entries, jobRefs, jobIDMap, nil
}

func writeWorkflowMetadata(sections *bodySections, doc ghworkflow.YAMLDocument, comments *yamlComments) error {
	body := sections.body
	appendSection := sections.next

	for _, key := range sortedKeys(doc.Raw) {
		if key == "jobs" {
			continue
		}

		appendSection()

		comment := comments.at(key)

		// Written here rather than inside each case: every key below becomes
		// either an attribute or a block, and the comment above it belongs
		// above whichever it becomes.
		hclcomment.WriteLeading(body, comment.head)

		value := doc.Raw[key]
		switch key {
		case "on":
			events := doc.On

			if len(events) == 0 {
				return errors.New("workflow 'on' must be an object")
			}

			onComments := comments.child(key)

			for i, eventName := range sortedKeys(events) {
				if i > 0 {
					appendSection()
				}

				// The comment above the whole "on" mapping is written by the
				// caller, above the first block it becomes. This is the one
				// above the event itself, which is the block being written.
				//
				// Both land above the same block when both were written, since
				// two YAML keys become one HCL block and there is nowhere else
				// for either to go. Stacked rather than dropped: a comment its
				// author wrote is worth more than a tidier line.
				hclcomment.WriteLeading(body, onComments.at(eventName).head)

				eventBlock := body.AppendNewBlock("on", []string{eventName})

				if err := writeOnEventBody(eventName, events[eventName], eventBlock.Body(), onComments.child(eventName)); err != nil {
					return err
				}
			}
		case "env":
			if err := writeNameValueBlocks(body, "env", value, comments.child(key)); err != nil {
				return err
			}
		case "permissions", "defaults", "concurrency":
			if err := writeNestedMapAsBlock(body, key, value, comments.child(key)); err != nil {
				return err
			}
		default:
			if err := writeCommentedAttribute(body, toHCLKey(key), value, comment.withoutHead()); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeWorkflowJobs(root *hclwrite.Body, jobs []workflowJobEntry, jobIDMap map[string]string, comments *yamlComments, generatedVariables map[string]any, stepRegistry map[string]string, usedStepIDs map[string]struct{}) error {
	jobComments := comments.child("jobs")

	for _, job := range jobs {
		jobName := job.Name
		jobMap := job.Body

		if len(root.Attributes()) > 0 || len(root.Blocks()) > 0 {
			root.AppendNewline()
		}

		hclcomment.WriteLeading(root, jobComments.at(jobName).head)

		jobID := jobIDMap[jobName]
		jobBlock := root.AppendNewBlock("job", []string{jobID})

		// Block labels are sanitized so the job can be referenced as
		// job.<id>, which turns "build-and-test" into "build_and_test". Keep
		// the original key as an attribute; parse prefers it when writing the
		// job back out.
		if jobName != jobID {
			jobBlock.Body().SetAttributeValue("id", cty.StringVal(jobName))
		}

		if err := writeJobBody(root, jobBlock.Body(), jobID, jobMap, jobIDMap, jobComments.child(jobName), generatedVariables, stepRegistry, usedStepIDs); err != nil {
			return fmt.Errorf("error in job '%s': %w", jobName, err)
		}
	}

	return nil
}

func writeGeneratedVariables(root *hclwrite.Body, generatedVariables map[string]any) error {
	if len(generatedVariables) == 0 {
		return nil
	}

	for _, varName := range sortedKeys(generatedVariables) {
		if len(root.Blocks()) > 0 || len(root.Attributes()) > 0 {
			root.AppendNewline()
		}

		vBlock := root.AppendNewBlock("variable", []string{varName})

		if err := writeAttributeAny(vBlock.Body(), "value", generatedVariables[varName]); err != nil {
			return err
		}
	}

	return nil
}

func writeJobKey(root *hclwrite.Body, body *hclwrite.Body, jobID string, key string, value any, jobIDMap map[string]string, comments *yamlComments, generatedVariables map[string]any, stepRegistry map[string]string, usedStepIDs map[string]struct{}, stepRefs *[]string) error {
	switch key {
	case "steps":
		refs, err := writeJobSteps(root, value, comments, stepRegistry, usedStepIDs)
		if err != nil {
			return err
		}
		*stepRefs = append(*stepRefs, refs...)

		return nil
	case "runs-on":
		return writeRunsOn(body, value, comments.child(key))
	case "needs":
		refs, err := normalizeNeeds(value, jobIDMap)
		if err != nil {
			return err
		}

		return writeReferenceListAttribute(body, "depends_on", "job", refs, "")
	case "env":
		return writeNameValueBlocks(body, "env", value, comments.child(key))
	case "with":
		return writeNameValueBlocks(body, "with", value, comments.child(key))
	case "outputs":
		return writeNameValueBlocks(body, "output", value, comments.child(key))
	case "services":
		return writeServicesBlocks(body, value)
	case "secrets":
		if str, ok := value.(string); ok {
			return writeAttributeAny(body, "secrets", str)
		}

		return writeNameValueBlocks(body, "secret", value, comments.child(key))
	case "strategy":
		return writeStrategyBlock(body, value, generatedVariables)
	case "permissions", "defaults", "concurrency", "container", "environment":
		return writeNestedMapAsBlock(body, key, value, comments.child(key))
	default:
		return writeCommentedAttribute(body, toHCLKey(key), value, comments.at(key).withoutHead())
	}
}

func writeJobSteps(root *hclwrite.Body, raw any, comments *yamlComments, stepRegistry map[string]string, usedStepIDs map[string]struct{}) ([]string, error) {
	items, ok := raw.([]any)

	if !ok {
		return nil, errors.New("job 'steps' must be a list")
	}

	stepRefs := make([]string, 0, len(items))

	for idx, item := range items {
		stepObj, ok := toStringAnyMap(item)

		if !ok {
			return nil, errors.New("job step must be an object")
		}

		if fp := stepFingerprint(stepObj, comments.item("steps", idx)); fp != "" {
			if existingID, exists := stepRegistry[fp]; exists {
				stepRefs = append(stepRefs, existingID)
				continue
			}
		}

		stepID := stepIdentifier(idx, stepObj, usedStepIDs)
		parsedStep, err := stepFromMap(stepObj, comments.item("steps", idx))
		if err != nil {
			return nil, err
		}

		parsedStep.Update(stepID)

		if err := parsedStep.Decode(root, "step"); err != nil {
			return nil, err
		}

		stepRegistry[stepFingerprint(stepObj, comments.item("steps", idx))] = stepID
		stepRefs = append(stepRefs, stepID)
	}

	return stepRefs, nil
}

func newWorkflowRoot(filename string) (*hclwrite.File, *hclwrite.Body, *hclwrite.Body) {
	f := hclwrite.NewEmptyFile()
	root := f.Body()

	workflowID := sanitizeIdentifier(filename)
	workflowBlock := root.AppendNewBlock("workflow", []string{workflowID})
	workflowBody := workflowBlock.Body()
	workflowBody.SetAttributeValue("filename", cty.StringVal(filename))

	return f, root, workflowBody
}
