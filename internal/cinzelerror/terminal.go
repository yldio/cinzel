// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package cinzelerror

import (
	"fmt"
	"strings"
)

// SafeForTerminal rewrites the control characters in s as escaped text.
//
// Errors quote names taken from the input file, and a name is free to carry an
// ANSI escape sequence. Written to a terminal as they are, those sequences are
// acted on rather than shown: a crafted job name can recolour the output, erase
// the line naming the file at fault, or move the cursor to forge a second
// message. Newline and tab are left alone because the YAML decoder quotes the
// offending source across several lines, and carriage return is not, because it
// returns to the start of a line already written.
//
// The bidirectional formatting characters go too. They are not controls, but
// they reorder what follows them for the rest of the line, which forges a
// message the same way a cursor move does: a job name ending in an override
// renders the text after it backwards, so the file named at fault is not the
// file shown. The zero-width joiner is left alone: an emoji sequence is built
// out of it, and cinzel already carries those through unchanged.
//
// Nothing else is touched, so a name in any language reads as written. A name
// that leans on a bidi mark to set its own direction reads with the mark shown
// rather than applied, which is worse to look at and still says what the name
// is.
func SafeForTerminal(s string) string {
	if !strings.ContainsFunc(s, isControl) {
		return s
	}

	var b strings.Builder

	b.Grow(len(s))

	for _, r := range s {
		if isControl(r) {
			_, _ = fmt.Fprintf(&b, "\\u%04x", r)

			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}

// isControl reports whether r is a character that must not reach a terminal
// unescaped: a C0 or C1 control, or one of the bidirectional formatting
// characters that reorder the text after them.
func isControl(r rune) bool {
	if r == '\n' || r == '\t' {
		return false
	}

	if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
		return true
	}

	switch r {
	// The embeddings and overrides, and the pop that closes them.
	case 0x202a, 0x202b, 0x202c, 0x202d, 0x202e:
		return true
	// The isolates, and the pop that closes them.
	case 0x2066, 0x2067, 0x2068, 0x2069:
		return true
	// The marks, which set the direction of the neutral characters by them.
	case 0x061c, 0x200e, 0x200f:
		return true
	}

	return false
}
