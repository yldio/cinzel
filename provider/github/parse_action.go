// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/internal/maputil"
	"github.com/yldio/cinzel/internal/naming"
)

// ActionYAMLFile pairs an action filename with its marshalled YAML content.
type ActionYAMLFile struct {
	Filename string
	Content  map[string]any
}

func parseHCLActions(actions []hclActionBlock, hv *hclparser.HCLVars, stepMap map[string]any) ([]ActionYAMLFile, error) {
	result := make([]ActionYAMLFile, 0, len(actions))

	for _, a := range actions {
		content, err := parseActionBody(a.Body, hv, stepMap)
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

func parseActionBody(body hcl.Body, hv *hclparser.HCLVars, stepMap map[string]any) (map[string]any, error) {
	sb, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil, errUnsupportedBodyType
	}

	out := make(map[string]any)

	// Parse attributes in sorted order for deterministic output.
	for _, name := range maputil.SortedKeys(sb.Attributes) {
		attr := sb.Attributes[name]
		switch name {
		case "filename":
			val, err := parseAttr(attr.Expr, hv)
			if err != nil {
				return nil, err
			}
			out["_filename"] = val
		default:
			val, err := parseAttr(attr.Expr, hv)
			if err != nil {
				return nil, err
			}
			out[naming.ToYAMLKey(name)] = val
		}
	}

	// Parse blocks.
	for _, block := range sb.Blocks {
		switch block.Type {
		case "input":
			if len(block.Labels) != 1 {
				return nil, errors.New("input block must have exactly one label")
			}

			inputMap, err := parseActionBlockAttrs(block.Body, hv)
			if err != nil {
				return nil, err
			}

			inputs := getOrCreateMap(out, "inputs")
			inputs[block.Labels[0]] = inputMap

		case "output":
			if len(block.Labels) != 1 {
				return nil, errors.New("output block must have exactly one label")
			}

			outputMap, err := parseActionBlockAttrs(block.Body, hv)
			if err != nil {
				return nil, err
			}

			outputs := getOrCreateMap(out, "outputs")
			outputs[block.Labels[0]] = outputMap

		case "runs":
			runsMap, err := parseActionRunsBlock(block.Body, hv, stepMap)
			if err != nil {
				return nil, err
			}

			out["runs"] = runsMap

		case "branding":
			brandingMap, err := parseActionBlockAttrs(block.Body, hv)
			if err != nil {
				return nil, err
			}

			out["branding"] = brandingMap

		default:
			child, err := parseActionBlockAttrs(block.Body, hv)
			if err != nil {
				return nil, err
			}

			if len(block.Labels) == 1 {
				mapping := getOrCreateMap(out, naming.ToYAMLKey(block.Type))
				mapping[block.Labels[0]] = child
			} else {
				out[naming.ToYAMLKey(block.Type)] = child
			}
		}
	}

	return out, nil
}

func parseActionRunsBlock(body hcl.Body, hv *hclparser.HCLVars, stepMap map[string]any) (map[string]any, error) {
	sb, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil, errUnsupportedBodyType
	}

	out := make(map[string]any)

	for _, name := range maputil.SortedKeys(sb.Attributes) {
		attr := sb.Attributes[name]
		if name == "steps" {
			// Steps are references: steps = [step.x, step.y]
			refs, err := parseReferenceList(attr.Expr, "step")
			if err != nil {
				return nil, err
			}

			steps := make([]any, 0, len(refs))
			for _, ref := range refs {
				stepVal, exists := stepMap[ref]
				if !exists {
					return nil, fmt.Errorf("cannot find step '%s'", ref)
				}
				steps = append(steps, stepVal)
			}

			out["steps"] = steps
			continue
		}

		val, err := parseAttr(attr.Expr, hv)
		if err != nil {
			return nil, err
		}

		out[naming.ToYAMLKey(name)] = val
	}

	// Handle nested blocks inside runs (e.g., env for docker actions).
	for _, block := range sb.Blocks {
		child, err := parseActionBlockAttrs(block.Body, hv)
		if err != nil {
			return nil, err
		}

		out[naming.ToYAMLKey(block.Type)] = child
	}

	return out, nil
}

func parseActionBlockAttrs(body hcl.Body, hv *hclparser.HCLVars) (map[string]any, error) {
	sb, ok := body.(*hclsyntax.Body)
	if !ok {
		return nil, errUnsupportedBlockBody
	}

	out := make(map[string]any)
	for _, name := range maputil.SortedKeys(sb.Attributes) {
		val, err := parseAttr(sb.Attributes[name].Expr, hv)
		if err != nil {
			return nil, err
		}
		out[naming.ToYAMLKey(name)] = val
	}

	return out, nil
}
