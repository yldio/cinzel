// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import "testing"

// stepIdentifier used to append "_N" from a counter without checking the
// result was free, so a document holding "build" twice plus a literal
// "build_1" produced two blocks labelled "build_1" and the output would not
// parse back.
func TestStepIdentifierNeverRepeatsALabel(t *testing.T) {
	for _, tc := range []struct {
		name  string
		steps []map[string]any
		want  []string
	}{
		{
			name:  "distinct names are untouched",
			steps: []map[string]any{{"id": "build"}, {"id": "test"}},
			want:  []string{"build", "test"},
		},
		{
			name:  "a repeat is suffixed",
			steps: []map[string]any{{"id": "build"}, {"id": "build"}},
			want:  []string{"build", "build_2"},
		},
		{
			name:  "a literal suffix is walked past",
			steps: []map[string]any{{"id": "build"}, {"id": "build_2"}, {"id": "build"}},
			want:  []string{"build", "build_2", "build_3"},
		},
		{
			// The label is lowercased, so two ids differing only in case
			// reach the same one.
			name:  "case does not make a label distinct",
			steps: []map[string]any{{"id": "Build"}, {"id": "build"}},
			want:  []string{"build", "build_2"},
		},
		{
			name:  "the positional fallback is claimed too",
			steps: []map[string]any{{"run": "  "}, {"id": "step_1"}},
			want:  []string{"step_1", "step_1_2"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			used := map[string]struct{}{}
			seen := map[string]struct{}{}

			for i, s := range tc.steps {
				got := stepIdentifier(i, s, used)

				if got != tc.want[i] {
					t.Errorf("step %d = %q, want %q", i, got, tc.want[i])
				}

				if _, repeat := seen[got]; repeat {
					t.Errorf("label %q was handed out twice", got)
				}
				seen[got] = struct{}{}
			}
		})
	}
}
