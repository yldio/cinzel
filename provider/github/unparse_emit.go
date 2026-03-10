// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
	"github.com/zclconf/go-cty/cty"
)

type workflowJobEntry struct {
	Name string
	Body map[string]any
}

func buildWorkflowJobIndex(jobs map[string]any) ([]workflowJobEntry, []string, map[string]string, error) {
	jobNames := sortedKeys(jobs)
	entries := make([]workflowJobEntry, 0, len(jobs))
	jobRefs := make([]string, 0, len(jobs))
	jobIDMap := make(map[string]string, len(jobs))

	for _, jobName := range jobNames {
		jobMap, ok := toStringAnyMap(jobs[jobName])

		if !ok {
			return nil, nil, nil, fmt.Errorf("job '%s' must be an object", jobName)
		}

		entries = append(entries, workflowJobEntry{Name: jobName, Body: jobMap})

		jobID := sanitizeIdentifier(jobName)

		if jobID == "" {
			jobID = "job"
		}

		jobID = uniqueIdentifier(jobID, jobRefs)
		jobRefs = append(jobRefs, jobID)
		jobIDMap[jobName] = jobID
	}

	return entries, jobRefs, jobIDMap, nil
}

func writeWorkflowMetadata(body *hclwrite.Body, doc ghworkflow.YAMLDocument) error {
	appendSection := func() {
		if len(body.Attributes()) > 0 || len(body.Blocks()) > 0 {
			body.AppendNewline()
		}
	}

	for _, key := range sortedKeys(doc.Raw) {
		if key == "jobs" {
			continue
		}

		appendSection()

		value := doc.Raw[key]
		switch key {
		case "on":
			events := doc.On

			if len(events) == 0 {
				return errors.New("workflow 'on' must be an object")
			}

			for _, eventName := range sortedKeys(events) {
				eventBlock := body.AppendNewBlock("on", []string{eventName})

				if err := writeOnEventBody(eventName, events[eventName], eventBlock.Body()); err != nil {
					return err
				}
			}
		case "env":
			if err := writeNameValueBlocks(body, "env", value); err != nil {
				return err
			}
		case "permissions", "defaults", "concurrency":
			if err := writeNestedMapAsBlock(body, key, value); err != nil {
				return err
			}
		default:
			if err := writeAttributeAny(body, toHCLKey(key), value); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeWorkflowJobs(root *hclwrite.Body, jobs []workflowJobEntry, jobIDMap map[string]string, generatedVariables map[string]any) error {
	for _, job := range jobs {
		jobName := job.Name
		jobMap := job.Body

		if len(root.Attributes()) > 0 || len(root.Blocks()) > 0 {
			root.AppendNewline()
		}

		jobID := jobIDMap[jobName]
		jobBlock := root.AppendNewBlock("job", []string{jobID})

		if err := writeJobBody(root, jobBlock.Body(), jobID, jobMap, jobIDMap, generatedVariables); err != nil {
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

func writeJobKey(root *hclwrite.Body, body *hclwrite.Body, jobID string, key string, value any, jobIDMap map[string]string, generatedVariables map[string]any, stepRefs *[]string) error {
	switch key {
	case "steps":
		refs, err := writeJobSteps(root, jobID, value)
		if err != nil {
			return err
		}
		*stepRefs = append(*stepRefs, refs...)

		return nil
	case "runs-on":
		return writeRunsOn(body, value)
	case "needs":
		refs, err := normalizeNeeds(value, jobIDMap)
		if err != nil {
			return err
		}

		return writeReferenceListAttribute(body, "depends_on", "job", refs)
	case "env":
		return writeNameValueBlocks(body, "env", value)
	case "with":
		return writeNameValueBlocks(body, "with", value)
	case "outputs":
		return writeNameValueBlocks(body, "output", value)
	case "services":
		return writeServicesBlocks(body, value)
	case "secrets":
		if str, ok := value.(string); ok {
			return writeAttributeAny(body, "secrets", str)
		}

		return writeNameValueBlocks(body, "secret", value)
	case "strategy":
		return writeStrategyBlock(body, value, generatedVariables)
	case "permissions", "defaults", "concurrency", "container", "environment":
		return writeNestedMapAsBlock(body, key, value)
	default:
		return writeAttributeAny(body, toHCLKey(key), value)
	}
}

func writeJobSteps(root *hclwrite.Body, jobID string, raw any) ([]string, error) {
	items, ok := raw.([]any)

	if !ok {
		return nil, errors.New("job 'steps' must be a list")
	}

	used := map[string]int{}
	stepRefs := make([]string, 0, len(items))

	for idx, item := range items {
		stepObj, ok := toStringAnyMap(item)

		if !ok {
			return nil, errors.New("job step must be an object")
		}

		stepID := stepIdentifier(jobID, idx, stepObj, used)
		parsedStep, err := stepFromMap(stepObj)
		if err != nil {
			return nil, err
		}

		parsedStep.Update(stepID)

		if err := parsedStep.Decode(root, "step"); err != nil {
			return nil, err
		}

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
