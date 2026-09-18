// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
)

// labelledBlockComments returns the comments written above and at the end of
// each block, keyed by block label.
func labelledBlockComments(hv *hclparser.HCLVars, blocks []*hcl.Block) map[string]blockComment {
	if len(blocks) == 0 {
		return nil
	}

	comments := map[string]blockComment{}

	for _, block := range blocks {
		// The head is read from the block header rather than the body,
		// because a header is what survives the merge of every file in the
		// directory with its source range intact.
		c := blockComment{
			head: hv.HeadComment(block.DefRange),
			foot: blockFootComment(block.Body, hv),
		}

		if c.head == "" && c.foot == "" {
			continue
		}

		comments[block.Labels[0]] = c
	}

	return comments
}

// labelledBlocks returns the headers of every block of the given type, which
// is where the source ranges survive the merge of every file in the directory.
//
// A header rather than a "remain" body, because a body that collects what the
// schema did not take also collects an attribute nobody declared, and an
// unknown attribute stops being an error.
func labelledBlocks(body hcl.Body, blockType string) []*hcl.Block {
	content, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: blockType, LabelNames: []string{"id"}}},
	})

	// Left to the full decode, which reports it with the detail this pass
	// deliberately does not collect.
	if diags.HasErrors() {
		return nil
	}

	return content.Blocks
}
