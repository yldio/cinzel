// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/hclcomment"
	ghjob "github.com/yldio/cinzel/provider/github/job"
	"github.com/zclconf/go-cty/cty"
)

func writeStrategyBlock(body *hclwrite.Body, raw any, comments *yamlComments, generatedVariables map[string]any) error {
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
			if err := writeMatrixBlock(strategyBody, value, comments.child(key), generatedVariables); err != nil {
				return err
			}
			continue
		}

		if err := checkHCLKeyRoundtrips("strategy", key); err != nil {
			return err
		}

		if err := writeCommentedAttribute(strategyBody, toHCLKey(key), value, comments.at(key)); err != nil {
			return err
		}
	}

	return nil
}

func writeMatrixBlock(body *hclwrite.Body, raw any, comments *yamlComments, generatedVariables map[string]any) error {
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

		comment := comments.at(axis.Name)

		if axis.Name == "include" || axis.Name == "exclude" {
			if err := writeCommentedAttribute(matrixBody, toHCLKey(axis.Name), axis.Value, comment); err != nil {
				return err
			}
			continue
		}

		if list, ok := axis.Value.([]any); ok {
			varName := uniqueIdentifierInSet("list_"+sanitizeIdentifier(axis.Name), existingNames)
			existingNames[varName] = struct{}{}
			generatedVariables[varName] = list

			// The axis is one YAML key and the block it becomes has two
			// attributes, so the comment goes on the value: that is the line
			// the axis list was written on, and where the parse reads it back.
			hclcomment.WriteLeading(matrixBody, comment.head)

			vBlock := matrixBody.AppendNewBlock("variable", nil)
			vBody := vBlock.Body()
			vBody.SetAttributeValue("name", cty.StringVal(axis.Name))
			vBody.SetAttributeRaw("value", hclcomment.Trailing(traversalTokens("variable", varName), comment.line))
			continue
		}

		if err := checkHCLIdentifier("matrix", axis.Name, axis.Name); err != nil {
			return err
		}

		if err := writeCommentedAttribute(matrixBody, axis.Name, axis.Value, comment); err != nil {
			return err
		}
	}

	return nil
}
