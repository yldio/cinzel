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

// lineComment is one whole-line comment, held with the line it starts on so a
// run can be walked past a comment that spans several.
type lineComment struct {
	text  string
	start int
}

// commentRunAbove returns the unbroken run of whole-line comments ending on the
// line above line.
func commentRunAbove(tokens []hclsyntax.Token, line int) string {
	if len(tokens) == 0 {
		return ""
	}

	lines := map[int]lineComment{}

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

		text := asYAMLComment(strings.TrimRight(string(tok.Bytes), "\n"))
		start := tok.Range.Start.Line

		// Keyed by the line the text ends on, because the run above is walked
		// upwards a line at a time. Range.End is no help: a "#" comment
		// swallows its own newline and so ends on the line below its text,
		// while a block comment ends exactly where it looks like it does.
		lines[start+strings.Count(text, "\n")] = lineComment{text: text, start: start}
	}

	var run []string

	for above := line - 1; ; {
		comment, found := lines[above]

		if !found {
			break
		}

		run = append([]string{comment.text}, run...)

		// A block comment covers every line between its own start and end, and
		// none of them is a key. Stepping one line at a time would find
		// nothing there and end a run that carries on above it.
		above = comment.start - 1
	}

	if len(run) == 0 {
		return ""
	}

	return strings.Join(run, "\n")
}

// asYAMLComment swaps an HCL comment marker for the one YAML uses, leaving the
// prose after it byte for byte.
//
// HCL writes a comment three ways and YAML has only "#", so a "// x" carried
// across verbatim arrives as "# // x": the marker read as part of the text.
// The marker is syntax rather than something its author wrote, so it is the
// one part of a comment that is translated.
//
// A block comment keeps its "/*" and "*/". Those wrap text that can run over
// several lines and have no YAML equivalent to swap in, and a reader seeing
// them knows what they came from.
func asYAMLComment(text string) string {
	trimmed := strings.TrimSpace(text)

	if !strings.HasPrefix(trimmed, "//") {
		return text
	}

	indent := text[:len(text)-len(strings.TrimLeft(text, " \t"))]

	return indent + "#" + strings.TrimPrefix(trimmed, "//")
}
