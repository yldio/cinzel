// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	ghworkflow "github.com/yldio/cinzel/provider/github/workflow"
)

func writeOnEventBody(event string, raw any, body *hclwrite.Body) error {
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
			if err := writeLabeledBlocks(body, blockType, value); err != nil {
				return err
			}
			continue
		}

		if err := writeAttributeAny(body, toHCLKey(key), value); err != nil {
			return err
		}
	}

	return nil
}

func writeLabeledBlocks(body *hclwrite.Body, blockType string, raw any) error {
	items, ok := toStringAnyMap(raw)

	if !ok {
		return fmt.Errorf("%s must be an object", blockType)
	}

	for _, label := range sortedKeys(items) {
		child := body.AppendNewBlock(blockType, []string{label})
		childBody := child.Body()

		childMap, ok := toStringAnyMap(items[label])

		if !ok {
			return fmt.Errorf("%s '%s' must be an object", blockType, label)
		}

		for _, key := range sortedKeys(childMap) {
			if err := writeAttributeAny(childBody, toHCLKey(key), childMap[key]); err != nil {
				return err
			}
		}
	}

	return nil
}
