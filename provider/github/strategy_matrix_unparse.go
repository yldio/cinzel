// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	ghjob "github.com/yldio/cinzel/provider/github/job"
	"github.com/zclconf/go-cty/cty"
)

func writeStrategyBlock(body *hclwrite.Body, raw any, generatedVariables map[string]any) error {
	mapping, ok := toStringAnyMap(raw)
	if !ok {
		return writeAttributeAny(body, "strategy", raw)
	}

	strategyBlock := body.AppendNewBlock("strategy", nil)
	strategyBody := strategyBlock.Body()

	for _, key := range sortedKeys(mapping) {
		if len(strategyBody.Attributes()) > 0 || len(strategyBody.Blocks()) > 0 {
			strategyBody.AppendNewline()
		}

		value := mapping[key]
		if key == "matrix" {
			if err := writeMatrixBlock(strategyBody, value, generatedVariables); err != nil {
				return err
			}
			continue
		}

		if err := writeAttributeAny(strategyBody, toHCLKey(key), value); err != nil {
			return err
		}
	}

	return nil
}

func writeMatrixBlock(body *hclwrite.Body, raw any, generatedVariables map[string]any) error {
	mapping, ok := toStringAnyMap(raw)
	if !ok {
		return writeAttributeAny(body, "matrix", raw)
	}

	matrixBlock := body.AppendNewBlock("matrix", nil)
	matrixBody := matrixBlock.Body()
	existingNames := make(map[string]struct{}, len(generatedVariables))
	for name := range generatedVariables {
		existingNames[name] = struct{}{}
	}

	axes := ghjob.AxesFromMap(mapping)
	for _, axis := range axes {
		if len(matrixBody.Attributes()) > 0 || len(matrixBody.Blocks()) > 0 {
			matrixBody.AppendNewline()
		}

		if axis.Name == "include" || axis.Name == "exclude" {
			if err := writeAttributeAny(matrixBody, toHCLKey(axis.Name), axis.Value); err != nil {
				return err
			}
			continue
		}

		if list, ok := axis.Value.([]any); ok {
			varName := uniqueIdentifierInSet("list_"+sanitizeIdentifier(axis.Name), existingNames)
			existingNames[varName] = struct{}{}
			generatedVariables[varName] = list

			vBlock := matrixBody.AppendNewBlock("variable", nil)
			vBody := vBlock.Body()
			vBody.SetAttributeValue("name", cty.StringVal(axis.Name))
			vBody.SetAttributeRaw("value", traversalTokens("variable", varName))
			continue
		}

		if err := writeAttributeAny(matrixBody, axis.Name, axis.Value); err != nil {
			return err
		}
	}

	return nil
}
