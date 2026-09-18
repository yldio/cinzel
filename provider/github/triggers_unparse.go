// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/hclcomment"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
)

func writeOnEventBody(event string, raw any, body *hclwrite.Body, comments *yamlComments) error {
	if raw == nil {
		return nil
	}

	eventMap, ok := toStringAnyMap(raw)

	if !ok {
		return writeAttributeAny(body, toHCLKey(event), raw)
	}

	for _, key := range sortedKeys(eventMap) {
		value := eventMap[key]

		if blockType, ok := ghworkflow.TriggerBlockTypeForEventKey(event, key); ok {
			// The key becomes one block per entry under it, so the comment
			// above the key goes above the first of them.
			hclcomment.WriteLeading(body, comments.at(key).head)

			if err := writeLabeledBlocks(body, blockType, value, comments.child(key)); err != nil {
				return err
			}
			continue
		}

		// A list value closes with a comment of its own, which is kept under
		// the key because a sequence has no node to carry it. The attribute
		// the list became is one line, so that comment closes the body.
		comment := comments.at(key)

		if err := writeCommentedAttribute(body, toHCLKey(key), value, comment); err != nil {
			return err
		}

		hclcomment.WriteLeading(body, comments.child(key).below())
	}

	hclcomment.WriteLeading(body, comments.below())

	return nil
}

func writeLabeledBlocks(body *hclwrite.Body, blockType string, raw any, comments *yamlComments) error {
	items, ok := toStringAnyMap(raw)

	if !ok {
		return fmt.Errorf("%s must be an object", blockType)
	}

	for _, label := range sortedKeys(items) {
		hclcomment.WriteLeading(body, comments.at(label).head)

		child := body.AppendNewBlock(blockType, []string{label})
		childBody := child.Body()

		childMap, ok := toStringAnyMap(items[label])

		if !ok {
			return fmt.Errorf("%s '%s' must be an object", blockType, label)
		}

		labelComments := comments.child(label)

		for _, key := range sortedKeys(childMap) {
			if err := writeCommentedAttribute(childBody, toHCLKey(key), childMap[key], labelComments.at(key)); err != nil {
				return err
			}

			hclcomment.WriteLeading(childBody, labelComments.child(key).below())
		}

		hclcomment.WriteLeading(childBody, labelComments.below())
	}

	hclcomment.WriteLeading(body, comments.below())

	return nil
}
