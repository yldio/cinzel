// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// HeadComment returns the run of whole-line comments written directly above r,
// as YAML wants it: each line still carrying its "#", joined by newlines.
//
// The run is unbroken. A blank line between the comment and what it sits above
// ends it, the same way it reads on the page, and a comment sharing a line with
// code is not part of it: that one belongs to whatever it trails, and taking it
// would move it somewhere it was not written.
//
// A comment is not part of any decoded value, so a lex over the source is the
// only way back to one.
func (av *HCLVars) HeadComment(r hcl.Range) string {
	return commentRunAbove(av.tokens(r.Filename), r.Start.Line)
}

// tokens returns the lexed tokens of one source file, lexing it on first ask
// and holding the result for the rest of the parse.
//
// Anything that fails to lex yields no tokens rather than an error. The decode
// finds the syntax error and says where it is; a second complaint from here
// would be the same problem told twice, in a worse way.
func (av *HCLVars) tokens(filename string) []hclsyntax.Token {
	if cached, lexed := av.lexed[filename]; lexed {
		return cached
	}

	src, held := av.sources[filename]

	if !held {
		return nil
	}

	parsed, diags := hclsyntax.LexConfig(src, filename, hcl.InitialPos)

	if diags.HasErrors() {
		parsed = nil
	}

	if av.lexed == nil {
		av.lexed = map[string][]hclsyntax.Token{}
	}

	av.lexed[filename] = parsed

	return parsed
}

// commentRunAbove returns the unbroken run of whole-line comments ending on the
// line above line.
func commentRunAbove(tokens []hclsyntax.Token, line int) string {
	if len(tokens) == 0 {
		return ""
	}

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
