// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclcomment

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// WriteLeading emits comment as HCL comment lines above whatever is appended
// next. An empty comment writes nothing.
//
// The text is written as it was read. A comment is prose, and its spacing, its
// "#" count and its indentation are things its author chose, so rewriting them
// changes what was written for no gain. The one thing added is a missing "#":
// a line reaching HCL without one is not a comment but a syntax error in the
// generated file.
func WriteLeading(body *hclwrite.Body, comment string) {
	if comment == "" {
		return
	}

	tokens := hclwrite.Tokens{}

	for _, line := range strings.Split(comment, "\n") {
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenComment,
			Bytes: []byte(Line(line) + "\n"),
		})
	}

	body.AppendUnstructuredTokens(tokens)
}

// Trailing returns tokens with comment appended, so it lands on the same line
// as the value those tokens write. An empty comment returns tokens unchanged.
//
// No newline in the comment bytes: the attribute brings its own, and a second
// one leaves a blank line after every commented attribute.
func Trailing(tokens hclwrite.Tokens, comment string) hclwrite.Tokens {
	if comment == "" {
		return tokens
	}

	return append(tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenComment,
		Bytes: []byte(" " + Line(comment)),
	})
}

// Line returns line as an HCL comment, adding a "#" only if the line does not
// already carry one. Everything else about the text is left alone.
func Line(line string) string {
	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		return line
	}

	return "# " + line
}
