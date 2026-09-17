// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// jobHeadComments returns the comment written directly above each job block,
// keyed by block label.
//
// A comment is not part of any decoded value, so a lex over the source is the
// only way back to one. The run of comment lines ending on the line before the
// block header is the comment that belongs to it: a blank line between them
// breaks the run, the same way it reads on the page.
//
// Anything a file fails to lex is skipped rather than reported. The decode
// finds the syntax error and says where it is; a second complaint from here
// would be the same problem told twice, in a worse way.
func jobHeadComments(sources map[string][]byte, blocks []*hcl.Block) map[string]string {
	if len(blocks) == 0 {
		return nil
	}

	byFile := map[string][]hclsyntax.Token{}
	comments := map[string]string{}

	for _, block := range blocks {
		name := block.DefRange.Filename

		toks, lexed := byFile[name]

		if !lexed {
			src, held := sources[name]

			if !held {
				continue
			}

			parsed, diags := hclsyntax.LexConfig(src, name, hcl.InitialPos)

			if diags.HasErrors() {
				parsed = nil
			}

			byFile[name] = parsed
			toks = parsed
		}

		if text := commentRunAbove(toks, block.DefRange.Start.Line); text != "" {
			comments[block.Labels[0]] = text
		}
	}

	return comments
}

// commentRunAbove returns the unbroken run of whole-line comments ending on
// the line above line, as YAML wants it: each line still carrying its "#",
// joined by newlines.
//
// A comment sharing a line with code is not part of the run. It belongs to
// whatever it trails, and taking it would move it somewhere it was not
// written.
func commentRunAbove(tokens []hclsyntax.Token, line int) string {
	lines := map[int]string{}

	for i, tok := range tokens {
		if tok.Type != hclsyntax.TokenComment {
			continue
		}

		// Start lines, not end: a comment token carries its own newline, so
		// the token before a whole-line comment ends on that comment's line
		// and every comment would read as trailing something.
		if i > 0 && tokens[i-1].Range.Start.Line == tok.Range.Start.Line {
			continue
		}

		lines[tok.Range.Start.Line] = strings.TrimRight(string(tok.Bytes), "\n")
	}

	var run []string

	for above := line - 1; ; above-- {
		text, found := lines[above]

		if !found {
			break
		}

		run = append([]string{text}, run...)
	}

	if len(run) == 0 {
		return ""
	}

	return strings.Join(run, "\n")
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
