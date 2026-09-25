// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package unescape_test

import (
	"testing"

	"github.com/yldio/cinzel/internal/unescape"
)

// Parity says an escape is doubled, which only a double-quoted scalar does. A
// plain YAML scalar and an HCL heredoc leave a backslash single, so text the
// author wrote arrived at the parity of an escape and was decoded: a script
// line reading `printf \u00e9` went out as `printf é`, at exit 0, on a value
// nothing had escaped. The rune settles it instead, because what the two
// writers escape is narrow and known.
func TestTextThatLooksLikeAnEscapeIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
		want string
	}{
		{
			name: "a shell line in a plain scalar",
			src:  `  - printf \u00e9 now`,
			want: `  - printf \u00e9 now`,
		},
		{
			name: "a windows path",
			src:  `  path: C:\users\u00e9dir`,
			want: `  path: C:\users\u00e9dir`,
		},
		{
			name: "an escape a writer does emit is still decoded",
			src:  `  name: "\U0001F46E Lint"`,
			want: "  name: \"\U0001F46E Lint\"",
		},
		{
			name: "a zero width joiner is still decoded",
			src:  `  name: "a\u200db"`,
			want: "  name: \"a\u200db\"",
		},
		{
			name: "a doubled backslash stays text",
			src:  `  run: printf \\u00e9`,
			want: `  run: printf \\u00e9`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(unescape.Unicode([]byte(tc.src))); got != tc.want {
				t.Errorf("Unicode() = %q, want %q", got, tc.want)
			}
		})
	}
}
