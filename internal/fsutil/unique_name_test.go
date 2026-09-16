// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import "testing"

func TestUniqueOutputName(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "distinct names are untouched",
			in:   []string{"ci", "release"},
			want: []string{"ci", "release"},
		},
		{
			name: "a repeat is suffixed",
			in:   []string{"ci", "ci", "ci"},
			want: []string{"ci", "ci_2", "ci_3"},
		},
		{
			// Folding decides the clash; the name written keeps the
			// casing the caller asked for.
			name: "case is folded when comparing",
			in:   []string{"ci", "CI"},
			want: []string{"ci", "CI_2"},
		},
		{
			name: "a suffix already taken is walked past",
			in:   []string{"ci", "ci_2", "ci"},
			want: []string{"ci", "ci_2", "ci_3"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			taken := map[string]struct{}{}

			for i, in := range tc.in {
				if got := UniqueOutputName(taken, in); got != tc.want[i] {
					t.Errorf("UniqueOutputName(%q) = %q, want %q", in, got, tc.want[i])
				}
			}
		})
	}
}
