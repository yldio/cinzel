// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
)

// ActionYAMLFile pairs an action filename with its marshalled YAML content.
type ActionYAMLFile struct {
	Filename string
	Content  map[string]any
	// FootComment closes the document, below its last key.
	FootComment string
}

func parseHCLActions(actions []hclActionBlock, body hcl.Body, hv *hclparser.HCLVars, stepMap map[string]any) ([]ActionYAMLFile, error) {
	result := make([]ActionYAMLFile, 0, len(actions))
	takenFilenames := make(map[string]string, len(actions))
	comments := labelledBlockComments(hv, labelledBlocks(body, "action"))

	for _, a := range actions {
		content, filename, err := parseActionConfig(a, hv, stepMap)
		if err != nil {
			return nil, fmt.Errorf("error in action '%s': %w", a.ID, err)
		}

		if filename == "" {
			filename = "action"
		}

		if err := checkFilenameStaysInside(filename); err != nil {
			return nil, fmt.Errorf("error in action '%s': %w", a.ID, err)
		}

		if err := claimFilename(takenFilenames, filename, a.ID); err != nil {
			return nil, err
		}

		result = append(result, ActionYAMLFile{
			Filename:    filename,
			Content:     content,
			FootComment: comments[a.ID].foot,
		})
	}

	return result, nil
}

// parseActionConfig converts one action block, returning its content and the
// filename to write it under.
func parseActionConfig(cfg hclActionBlock, hv *hclparser.HCLVars, stepMap map[string]any) (map[string]any, string, error) {
	out := make(map[string]any)

	filename, err := parseAttr(cfg.Filename, hv)
	if err != nil {
		return nil, "", err
	}

	name, _ := filename.(string)

	if err := setOptionalYAMLAttr(out, "name", cfg.Name, hv); err != nil {
		return nil, "", err
	}

	if err := setOptionalYAMLAttr(out, "description", cfg.Description, hv); err != nil {
		return nil, "", err
	}

	if err := setOptionalYAMLAttr(out, "author", cfg.Author, hv); err != nil {
		return nil, "", err
	}

	for _, input := range cfg.Inputs {
		inputMap := make(map[string]any)

		if err := setOptionalYAMLAttr(inputMap, "description", input.Description, hv); err != nil {
			return nil, "", err
		}

		if err := setOptionalYAMLAttr(inputMap, "required", input.Required, hv); err != nil {
			return nil, "", err
		}

		if err := setOptionalYAMLAttr(inputMap, "default", input.Default, hv); err != nil {
			return nil, "", err
		}

		// GitHub spells this one in camel case, alone among the input keys.
		// Writing "deprecation-message" produced a file GitHub ignores and
		// cinzel's own strict shape rejects on the way back.
		if err := setOptionalYAMLAttr(inputMap, "deprecationMessage", input.DeprecationMessage, hv); err != nil {
			return nil, "", err
		}

		inputs := getOrCreateMap(out, "inputs")

		if err := claimBlockKey(inputs, "input", input.ID); err != nil {
			return nil, "", err
		}

		inputs[input.ID] = inputMap
	}

	for _, output := range cfg.Outputs {
		outputMap := make(map[string]any)

		if err := setOptionalYAMLAttr(outputMap, "description", output.Description, hv); err != nil {
			return nil, "", err
		}

		if err := setOptionalYAMLAttr(outputMap, "value", output.Value, hv); err != nil {
			return nil, "", err
		}

		outputs := getOrCreateMap(out, "outputs")

		if err := claimBlockKey(outputs, "output", output.ID); err != nil {
			return nil, "", err
		}

		outputs[output.ID] = outputMap
	}

	for _, runs := range cfg.Runs {
		runsMap, err := parseActionRunsConfig(runs, hv, stepMap)
		if err != nil {
			return nil, "", err
		}

		if err := claimSingletonBlock(out, "runs", "runs"); err != nil {
			return nil, "", err
		}

		out["runs"] = runsMap
	}

	for _, branding := range cfg.Branding {
		brandingMap := make(map[string]any)

		if err := setOptionalYAMLAttr(brandingMap, "icon", branding.Icon, hv); err != nil {
			return nil, "", err
		}

		if err := setOptionalYAMLAttr(brandingMap, "color", branding.Color, hv); err != nil {
			return nil, "", err
		}

		if err := claimSingletonBlock(out, "branding", "branding"); err != nil {
			return nil, "", err
		}

		out["branding"] = brandingMap
	}

	return out, name, nil
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

	if err := setOptionalYAMLAttr(out, "pre-entrypoint", cfg.PreEntrypoint, hv); err != nil {
		return nil, err
	}

	if err := setOptionalYAMLAttr(out, "post-entrypoint", cfg.PostEntrypoint, hv); err != nil {
		return nil, err
	}

	if refs, err := parseReferenceList(cfg.Steps, "step"); err != nil {
		return nil, err
	} else if len(refs) > 0 {
		steps := make([]any, 0, len(refs))
		emitted := make(map[string]struct{}, len(refs))

		// GitHub applies the same step id uniqueness rule inside a composite
		// action as it does inside a job, so the guard the workflow path keeps
		// is kept here too. Without it an action went out as a written file and
		// exit 0, and only actionlint or GitHub itself said otherwise.
		takenIDs := make(map[string]string, len(refs))

		for _, ref := range refs {
			stepVal, exists := stepMap[ref]

			if !exists {
				return nil, fmt.Errorf("cannot find step '%s'", ref)
			}

			// An action may run the same step more than once. Only the first
			// occurrence carries the id.
			if _, repeat := emitted[ref]; repeat {
				stepVal = stepValueWithoutID(stepVal)
			} else if id, ok := emittedStepID(stepVal); ok {
				if other, taken := takenIDs[id]; taken {
					return nil, fmt.Errorf("%w: '%s' and '%s' both write '%s'",
						errDuplicateActionStepID, other, ref, id)
				}

				takenIDs[id] = ref
			}

			if err := checkStepNotEmpty(stepVal, ref); err != nil {
				return nil, err
			}

			emitted[ref] = struct{}{}

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

		if err := claimBlockKey(envMap, "env", key); err != nil {
			return nil, err
		}

		envMap[key] = withComments(value, namedBlockComments(env, hv))
	}

	return out, nil
}
