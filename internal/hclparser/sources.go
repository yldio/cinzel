// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"bytes"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// SetSources hands the store the bytes each file was parsed from, keyed by
// filename, as hclparse.Parser.Sources returns them.
//
// A comment is not part of any decoded value, so the only way back to one is
// the source text. The parser already holds every byte it read, and the store
// is the one thing threaded to each Parse along the way, so it carries them
// rather than a second argument added to every signature between here and the
// attribute.
func (av *HCLVars) SetSources(sources map[string][]byte) {
	av.sources = sources
}

// TrailingComment returns the # comment sharing a line with the end of r, or
// empty string if the line ends without one.
//
// Reading from what was parsed rather than from disk keeps the comment matched
// to the text the rest of the parse ran against, and costs one map lookup where
// re-reading cost a syscall per attribute.
func (av *HCLVars) TrailingComment(r hcl.Range) string {
	src, ok := av.sources[r.Filename]

	if !ok || int(r.End.Byte) >= len(src) {
		return ""
	}

	rest := src[r.End.Byte:]
	newline := bytes.IndexByte(rest, '\n')

	if newline < 0 {
		newline = len(rest)
	}

	tail := strings.TrimSpace(string(rest[:newline]))

	if !strings.HasPrefix(tail, "#") {
		return ""
	}

	return tail
}
