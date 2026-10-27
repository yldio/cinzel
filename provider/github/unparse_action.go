// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"
	"github.com/yldio/cinzel/internal/hclcomment"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/unescape"
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
// definition (has "runs" but not "on" or "jobs").
//
// A "name" is no longer part of the test. It is required of an action and
// validateActionDocument says so, but requiring it here sent a nameless action
// down the step-only path, which refused it with "not a valid type" and no
// mention of the field that was missing.
func isActionDocument(doc map[string]any) bool {
	_, hasRuns := doc["runs"]
	_, hasOn := doc["on"]
	_, hasJobs := doc["jobs"]

	return hasRuns && !hasOn && !hasJobs
}

// actionToHCL converts one action document. location is where the action was
// read from, relative to the run's root: "actions-a/setup" for an action at
// "actions-a/setup/action.yml".
//
// The last segment is the action's name and the rest is where it lives, and
// the two are written apart. One string did both jobs before, which worked
// only while an action could not sit anywhere but the output directory: the
// block label comes from sanitizeIdentifier, and a path through that becomes
// "actions_a_setup" rather than "setup".
func actionToHCL(doc map[string]any, location string, comments *yamlComments, site commentSite, lines map[string]int, ignoreUnknown bool, usedStepIDs map[string]struct{}) ([]byte, error) {
	if ignoreUnknown {
		reportDroppedKeys(site, stripUnknownKeys(doc, actionYAMLShape{}, lines))
	}

	if err := validateActionDocument(doc); err != nil {
		return nil, err
	}

	f := hclwrite.NewEmptyFile()
	root := f.Body()

	actionPath, filename := path.Split(path.Clean(filepath.ToSlash(location)))
	actionPath = strings.TrimSuffix(actionPath, "/")

	actionID := sanitizeIdentifier(filename)

	if actionID == "" {
		actionID = "action"
	}

	// Write step blocks for composite actions.
	var stepRefs []string

	if runs, ok := toStringAnyMap(doc["runs"]); ok {
		if using, _ := runs["using"].(string); using == "composite" {
			if stepsRaw, ok := runs["steps"]; ok {
				refs, err := writeActionSteps(root, site.child("runs").child("steps"), stepsRaw, comments.child("runs"), usedStepIDs)
				if err != nil {
					return nil, err
				}
				stepRefs = refs
			}
		}
	}

	// A blank line between the steps and the block that uses them. Each step
	// closes with one, so writing the action straight on put "action" hard
	// against the "}" above it and the file read as one run of blocks.
	//
	// Asked of the body rather than counted: a file with no steps in it has
	// nothing above the action to separate it from.
	if len(stepRefs) > 0 {
		root.AppendNewline()
	}

	actionBlock := root.AppendNewBlock("action", []string{actionID})
	actionBody := actionBlock.Body()

	actionBody.SetAttributeValue("filename", cty.StringVal(filename))

	// Only written when there is one. An action directly under the output
	// directory has nothing to say about where it sits, and every file
	// converted before this attribute existed looks exactly like that.
	if actionPath != "" && actionPath != "." {
		actionBody.SetAttributeValue("path", cty.StringVal(actionPath))
	}

	actionSections := newBodySections(actionBody)

	// Write top-level attributes in a stable order.

	for _, key := range sortedKeys(doc) {
		switch key {
		case "runs", "inputs", "outputs", "branding":
			continue // handled as blocks below
		default:
			actionSections.next()

			// With its comments, the way the workflow writer does it. Written
			// through writeAttributeAny, which carries none, the note above
			// an action's first key — the one describing the whole file —
			// was read off the YAML and then dropped on the floor.
			if err := writeCommentedAttribute(actionBody, toHCLKey(key), doc[key], comments.at(key)); err != nil {
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
			actionSections.next()

			inputMap, ok := toStringAnyMap(inputs[name])

			if !ok {
				return nil, fmt.Errorf("action input '%s' must be an object", name)
			}

			inputComments := comments.child("inputs")

			hclcomment.WriteLeading(actionBody, inputComments.at(name).head)

			inputBlock := actionBody.AppendNewBlock("input", []string{name})
			inputBody := inputBlock.Body()
			attrComments := inputComments.child(name)

			for _, attr := range sortedKeys(inputMap) {
				if err := writeCommentedAttribute(inputBody, inputAttrHCLKey(attr), inputMap[attr], attrComments.at(attr)); err != nil {
					return nil, err
				}
			}

			hclcomment.WriteLeading(inputBody, attrComments.below())
		}
	}

	// Write output blocks.

	if outputsRaw, ok := doc["outputs"]; ok {
		outputs, mapOK := toStringAnyMap(outputsRaw)

		if !mapOK {
			return nil, errors.New("action 'outputs' must be an object")
		}

		for _, name := range sortedKeys(outputs) {
			actionSections.next()

			outputMap, ok := toStringAnyMap(outputs[name])

			if !ok {
				return nil, fmt.Errorf("action output '%s' must be an object", name)
			}

			outputComments := comments.child("outputs")

			hclcomment.WriteLeading(actionBody, outputComments.at(name).head)

			outputBlock := actionBody.AppendNewBlock("output", []string{name})
			outputBody := outputBlock.Body()
			attrComments := outputComments.child(name)

			for _, attr := range sortedKeys(outputMap) {
				if err := writeCommentedAttribute(outputBody, toHCLKey(attr), outputMap[attr], attrComments.at(attr)); err != nil {
					return nil, err
				}
			}

			hclcomment.WriteLeading(outputBody, attrComments.below())
		}
	}

	// Write runs block.

	if runsRaw, ok := doc["runs"]; ok {
		runsMap, mapOK := toStringAnyMap(runsRaw)

		if !mapOK {
			return nil, errors.New("action 'runs' must be an object")
		}

		actionSections.next()

		runsBlock := actionBody.AppendNewBlock("runs", nil)
		runsBody := runsBlock.Body()
		runsSections := newBodySections(runsBody)

		for _, key := range sortedKeys(runsMap) {
			if key == "steps" {
				// Steps are written as top-level step blocks, referenced here.
				continue
			}

			if key == "env" {
				if err := writeNameValueBlocks(runsBody, "env", runsMap[key], comments.child("runs").child(key)); err != nil {
					return nil, err
				}
				continue
			}

			runsSections.next()

			if err := writeAttributeAny(runsBody, toHCLKey(key), runsMap[key]); err != nil {
				return nil, err
			}
		}

		if len(stepRefs) > 0 {
			runsSections.next()

			if err := writeReferenceListAttribute(runsBody, "steps", "step", stepRefs, comments.child("runs").at("steps").head); err != nil {
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

		actionSections.next()

		brandingBlock := actionBody.AppendNewBlock("branding", nil)
		brandingBody := brandingBlock.Body()

		for _, attr := range sortedKeys(brandingMap) {
			if err := writeAttributeAny(brandingBody, toHCLKey(attr), brandingMap[attr]); err != nil {
				return nil, err
			}
		}
	}

	return unescape.Unicode(hclwrite.Format(f.Bytes())), nil
}

func writeActionSteps(root *hclwrite.Body, site commentSite, raw any, comments *yamlComments, used map[string]struct{}) ([]string, error) {
	items, ok := raw.([]any)

	if !ok {
		return nil, errors.New("action runs.steps must be a list")
	}

	stepRefs := make([]string, 0, len(items))

	for idx, item := range items {
		stepObj, ok := toStringAnyMap(item)

		if !ok {
			return nil, errors.New("action step must be an object")
		}

		stepID := stepIdentifier(idx, stepObj, used)
		parsedStep, err := stepFromMap(site.index(idx), stepObj, comments.item("steps", idx))
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

// inputAttrHCLKey names an action input attribute in HCL. The general rule
// only swaps hyphens for underscores, which leaves GitHub's lone camel-case
// key as a literal "deprecationMessage" attribute the schema does not declare,
// so the emitted file did not parse back.
func inputAttrHCLKey(attr string) string {
	if attr == "deprecationMessage" {
		return "deprecation_message"
	}

	return toHCLKey(attr)
}
