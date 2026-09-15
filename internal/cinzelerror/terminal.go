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
// Only control characters are touched, so a name in any language reads as
// written.
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

// isControl reports whether r is a C0 or C1 control character that must not
// reach a terminal unescaped.
func isControl(r rune) bool {
	if r == '\n' || r == '\t' {
		return false
	}

	return r < 0x20 || (r >= 0x7f && r <= 0x9f)
}
