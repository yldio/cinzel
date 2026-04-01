// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// classifyActionDocument checks whether a pre-parsed YAML map looks like a
// GitHub Action definition. Returns the map if so, nil otherwise.
func classifyActionDocument(doc map[string]any) map[string]any {
	if isActionDocument(doc) {
		return doc
	}

	return nil
}

// isActionDocument returns true if the YAML document looks like a GitHub Action
// definition (has "name" and "runs" but not "on" or "jobs").
func isActionDocument(doc map[string]any) bool {
	_, hasRuns := doc["runs"]
	_, hasName := doc["name"]
	_, hasOn := doc["on"]
	_, hasJobs := doc["jobs"]

	return hasRuns && hasName && !hasOn && !hasJobs
}

func actionToHCL(doc map[string]any, filename string) ([]byte, error) {
	if err := validateActionDocument(doc); err != nil {
		return nil, err
	}

	f := hclwrite.NewEmptyFile()
	root := f.Body()

	actionID := sanitizeIdentifier(filename)

	if actionID == "" {
		actionID = "action"
	}

	// Write step blocks for composite actions.
	var stepRefs []string

	if runs, ok := toStringAnyMap(doc["runs"]); ok {
		if using, _ := runs["using"].(string); using == "composite" {
			if stepsRaw, ok := runs["steps"]; ok {
				refs, err := writeActionSteps(root, stepsRaw)
				if err != nil {
					return nil, err
				}
				stepRefs = refs
			}
		}
	}

	actionBlock := root.AppendNewBlock("action", []string{actionID})
	actionBody := actionBlock.Body()

	actionBody.SetAttributeValue("filename", cty.StringVal(filename))

	// Write top-level attributes in a stable order.

	for _, key := range sortedKeys(doc) {
		switch key {
		case "runs", "inputs", "outputs", "branding":
			continue // handled as blocks below
		default:
			if len(actionBody.Attributes()) > 0 || len(actionBody.Blocks()) > 0 {
				actionBody.AppendNewline()
			}

			if err := writeAttributeAny(actionBody, toHCLKey(key), doc[key]); err != nil {
				return nil, err
			}
		}
	}

	// Write input blocks.

	if inputsRaw, ok := doc["inputs"]; ok {
		inputs, mapOK := toStringAnyMap(inputsRaw)

		if !mapOK {
			return nil, errors.New("action 'inputs' must be an object")
		}

		for _, name := range sortedKeys(inputs) {
			if len(actionBody.Attributes()) > 0 || len(actionBody.Blocks()) > 0 {
				actionBody.AppendNewline()
			}

			inputMap, ok := toStringAnyMap(inputs[name])

			if !ok {
				return nil, fmt.Errorf("action input '%s' must be an object", name)
			}

			inputBlock := actionBody.AppendNewBlock("input", []string{name})
			inputBody := inputBlock.Body()

			for _, attr := range sortedKeys(inputMap) {
				if err := writeAttributeAny(inputBody, toHCLKey(attr), inputMap[attr]); err != nil {
					return nil, err
				}
			}
		}
	}

	// Write output blocks.

	if outputsRaw, ok := doc["outputs"]; ok {
		outputs, mapOK := toStringAnyMap(outputsRaw)

		if !mapOK {
			return nil, errors.New("action 'outputs' must be an object")
		}

		for _, name := range sortedKeys(outputs) {
			if len(actionBody.Attributes()) > 0 || len(actionBody.Blocks()) > 0 {
				actionBody.AppendNewline()
			}

			outputMap, ok := toStringAnyMap(outputs[name])

			if !ok {
				return nil, fmt.Errorf("action output '%s' must be an object", name)
			}

			outputBlock := actionBody.AppendNewBlock("output", []string{name})
			outputBody := outputBlock.Body()

			for _, attr := range sortedKeys(outputMap) {
				if err := writeAttributeAny(outputBody, toHCLKey(attr), outputMap[attr]); err != nil {
					return nil, err
				}
			}
		}
	}

	// Write runs block.

	if runsRaw, ok := doc["runs"]; ok {
		runsMap, mapOK := toStringAnyMap(runsRaw)

		if !mapOK {
			return nil, errors.New("action 'runs' must be an object")
		}

		if len(actionBody.Attributes()) > 0 || len(actionBody.Blocks()) > 0 {
			actionBody.AppendNewline()
		}

		runsBlock := actionBody.AppendNewBlock("runs", nil)
		runsBody := runsBlock.Body()

		for _, key := range sortedKeys(runsMap) {
			if key == "steps" {
				// Steps are written as top-level step blocks, referenced here.
				continue
			}

			if key == "env" {
				if err := writeNameValueBlocks(runsBody, "env", runsMap[key]); err != nil {
					return nil, err
				}
				continue
			}

			if len(runsBody.Attributes()) > 0 || len(runsBody.Blocks()) > 0 {
				runsBody.AppendNewline()
			}

			if err := writeAttributeAny(runsBody, toHCLKey(key), runsMap[key]); err != nil {
				return nil, err
			}
		}

		if len(stepRefs) > 0 {
			if len(runsBody.Attributes()) > 0 || len(runsBody.Blocks()) > 0 {
				runsBody.AppendNewline()
			}

			if err := writeReferenceListAttribute(runsBody, "steps", "step", stepRefs); err != nil {
				return nil, err
			}
		}
	}

	// Write branding block.

	if brandingRaw, ok := doc["branding"]; ok {
		brandingMap, mapOK := toStringAnyMap(brandingRaw)

		if !mapOK {
			return nil, errors.New("action 'branding' must be an object")
		}

		if len(actionBody.Attributes()) > 0 || len(actionBody.Blocks()) > 0 {
			actionBody.AppendNewline()
		}

		brandingBlock := actionBody.AppendNewBlock("branding", nil)
		brandingBody := brandingBlock.Body()

		for _, attr := range sortedKeys(brandingMap) {
			if err := writeAttributeAny(brandingBody, toHCLKey(attr), brandingMap[attr]); err != nil {
				return nil, err
			}
		}
	}

	return unescapeHCLUnicode(hclwrite.Format(f.Bytes())), nil
}

func writeActionSteps(root *hclwrite.Body, raw any) ([]string, error) {
	items, ok := raw.([]any)

	if !ok {
		return nil, errors.New("action runs.steps must be a list")
	}

	used := map[string]int{}
	stepRefs := make([]string, 0, len(items))

	for idx, item := range items {
		stepObj, ok := toStringAnyMap(item)

		if !ok {
			return nil, errors.New("action step must be an object")
		}

		stepID := stepIdentifier(idx, stepObj, used)
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

func validateActionDocument(doc map[string]any) error {
	if err := strictValidateYAMLShape(doc, &actionYAMLShape{}); err != nil {
		return err
	}

	if _, ok := doc["name"]; !ok {
		return errors.New("action must define 'name'")
	}

	runsRaw, ok := doc["runs"]

	if !ok {
		return errors.New("action must define 'runs'")
	}

	runs, ok := toStringAnyMap(runsRaw)

	if !ok {
		return errors.New("action 'runs' must be an object")
	}

	using, ok := runs["using"].(string)

	if !ok || using == "" {
		return errors.New("action 'runs.using' must be a non-empty string")
	}

	return nil
}
