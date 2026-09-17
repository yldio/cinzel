// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
)

// jobHeadComments returns the comment written directly above each job block,
// keyed by block label.
func jobHeadComments(hv *hclparser.HCLVars, blocks []*hcl.Block) map[string]string {
	if len(blocks) == 0 {
		return nil
	}

	comments := map[string]string{}

	for _, block := range blocks {
		if text := hv.HeadComment(block.DefRange); text != "" {
			comments[block.Labels[0]] = text
		}
	}

	return comments
}

// jobBlocks returns the job block headers, which is where the source ranges
// survive the merge of every file in the directory.
func jobBlocks(body hcl.Body) []*hcl.Block {
	content, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "job", LabelNames: []string{"id"}}},
	})

	// Left to the full decode, which reports it with the detail this pass
	// deliberately does not collect.
	if diags.HasErrors() {
		return nil
	}

	return content.Blocks
}
