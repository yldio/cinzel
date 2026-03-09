// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"fmt"

	"github.com/yldio/cinzel/internal/hclparser"
)

// ActionYAMLFile pairs an action filename with its marshalled YAML content.
type ActionYAMLFile struct {
	Filename string
	Content  map[string]any
}

func parseHCLActions(actions []hclActionBlock, hv *hclparser.HCLVars, stepMap map[string]any) ([]ActionYAMLFile, error) {
	result := make([]ActionYAMLFile, 0, len(actions))

	for _, a := range actions {
		content, err := parseActionConfig(a, hv, stepMap)
		if err != nil {

			return nil, fmt.Errorf("error in action '%s': %w", a.ID, err)
		}

		filename, _ := content["_filename"].(string)
		delete(content, "_filename")

		if filename == "" {
			filename = "action"
		}

		result = append(result, ActionYAMLFile{
			Filename: filename,
			Content:  content,
		})
	}

	return result, nil
}

func parseActionConfig(cfg hclActionBlock, hv *hclparser.HCLVars, stepMap map[string]any) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalYAMLAttr(out, "_filename", cfg.Filename, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "name", cfg.Name, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "description", cfg.Description, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "author", cfg.Author, hv); err != nil {

		return nil, err
	}

	for _, input := range cfg.Inputs {
		inputMap := make(map[string]any)

		if err := setOptionalYAMLAttr(inputMap, "description", input.Description, hv); err != nil {

			return nil, err
		}

		if err := setOptionalYAMLAttr(inputMap, "required", input.Required, hv); err != nil {

			return nil, err
		}

		if err := setOptionalYAMLAttr(inputMap, "default", input.Default, hv); err != nil {

			return nil, err
		}

		if err := setOptionalYAMLAttr(inputMap, "deprecation-message", input.DeprecationMessage, hv); err != nil {

			return nil, err
		}

		inputs := getOrCreateMap(out, "inputs")
		inputs[input.ID] = inputMap
	}

	for _, output := range cfg.Outputs {
		outputMap := make(map[string]any)

		if err := setOptionalYAMLAttr(outputMap, "description", output.Description, hv); err != nil {

			return nil, err
		}

		if err := setOptionalYAMLAttr(outputMap, "value", output.Value, hv); err != nil {

			return nil, err
		}

		outputs := getOrCreateMap(out, "outputs")
		outputs[output.ID] = outputMap
	}

	for _, runs := range cfg.Runs {
		runsMap, err := parseActionRunsConfig(runs, hv, stepMap)
		if err != nil {

			return nil, err
		}

		out["runs"] = runsMap
	}

	for _, branding := range cfg.Branding {
		brandingMap := make(map[string]any)

		if err := setOptionalYAMLAttr(brandingMap, "icon", branding.Icon, hv); err != nil {

			return nil, err
		}

		if err := setOptionalYAMLAttr(brandingMap, "color", branding.Color, hv); err != nil {

			return nil, err
		}

		out["branding"] = brandingMap
	}

	return out, nil
}

func parseActionRunsConfig(cfg hclActionRunsBlock, hv *hclparser.HCLVars, stepMap map[string]any) (map[string]any, error) {
	out := make(map[string]any)

	if err := setOptionalYAMLAttr(out, "using", cfg.Using, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "main", cfg.Main, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "pre", cfg.Pre, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "pre-if", cfg.PreIf, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "post", cfg.Post, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "post-if", cfg.PostIf, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "image", cfg.Image, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "args", cfg.Args, hv); err != nil {

		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "entrypoint", cfg.Entrypoint, hv); err != nil {

		return nil, err
	}

	if refs, err := parseReferenceList(cfg.Steps, "step"); err != nil {

		return nil, err
	} else if len(refs) > 0 {
		steps := make([]any, 0, len(refs))

		for _, ref := range refs {
			stepVal, exists := stepMap[ref]
			if !exists {

				return nil, fmt.Errorf("cannot find step '%s'", ref)
			}

			steps = append(steps, stepVal)
		}

		out["steps"] = steps
	}

	for _, env := range cfg.Env {
		key, value, err := parseNamedConfig(env, hv)
		if err != nil {

			return nil, err
		}

		envMap := getOrCreateMap(out, "env")
		envMap[key] = value
	}

	return out, nil
}
