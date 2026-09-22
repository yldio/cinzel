// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package cinzelerror

import (
	"strings"
	"testing"
)

// A bidirectional override reorders everything after it for the rest of the
// line, so a name carrying one rewrites the message that quotes it: the file
// named at fault renders as a different file. It is the forgery an ANSI cursor
// move already could not do, by another route.
func TestSafeForTerminalEscapesBidiOverrides(t *testing.T) {
	t.Parallel()

	tests := map[string]rune{
		"right-to-left override":  0x202e,
		"left-to-right override":  0x202d,
		"right-to-left embedding": 0x202b,
		"left-to-right embedding": 0x202a,
		"pop directional format":  0x202c,
		"right-to-left isolate":   0x2067,
		"left-to-right isolate":   0x2066,
		"first strong isolate":    0x2068,
		"pop directional isolate": 0x2069,
		"right-to-left mark":      0x200f,
		"left-to-right mark":      0x200e,
		"arabic letter mark":      0x061c,
	}

	for name, r := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			in := "job" + string(r) + "manifest.html"

			got := SafeForTerminal(in)
			if strings.ContainsRune(got, r) {
				t.Errorf("%U reached the terminal: %q", r, got)
			}
		})
	}
}

// The zero-width joiner holds an emoji sequence together, and cinzel carries
// those through unchanged everywhere else. Escaping it here would print an
// emoji in a workflow name as its pieces. Without this the switch could be
// widened to every format character and still look correct.
func TestSafeForTerminalKeepsTheJoinerAnEmojiIsBuiltFrom(t *testing.T) {
	t.Parallel()

	for _, in := range []string{
		"\U0001f46e‍♂️ Lint",
		"jobs.\U0001f3f4\U000e0067\U000e0062\U000e0073\U000e0063\U000e0074\U000e007f",
	} {
		if got := SafeForTerminal(in); got != in {
			t.Errorf("SafeForTerminal(%q) = %q, want it unchanged", in, got)
		}
	}
}
