// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// "jobs" and "steps" become top-level blocks referenced from a list, so the
// writers above skip them with a "continue" before the line that writes a head
// comment, and the emitter they reach took no comment at all. A note above
// either key was dropped on unparse, at exit 0, on the two keys most likely to
// carry one.
func TestAReferenceListKeepsTheCommentAboveIt(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "above the workflow jobs key",
			yaml: "on:\n  push: {}\n# what this builds\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n",
			want: "# what this builds\n  jobs = [",
		},
		{
			name: "above a job steps key",
			yaml: "on:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    # what it does\n    steps:\n      - run: make\n",
			want: "# what it does\n  steps = [",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl := unparse(t, tc.yaml)

			if !strings.Contains(hcl, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, hcl)
			}
		})
	}
}
