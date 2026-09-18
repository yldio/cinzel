// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// Three bodies on the way back from YAML took no comments at all: an "on"
// event, a name-value block such as "env", and "runs-on". Each is written by a
// helper that was never handed the comment tree, so every comment written
// inside one was dropped while the rest of the file kept its own.
func TestNestedBlockCommentsSurviveUnparse(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "above an event",
			yaml: "on:\n  # only main\n  push:\n    branches:\n      - main\n" +
				"jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n",
			want: "# only main\n  on \"push\"",
		},
		{
			name: "above an event attribute",
			yaml: "on:\n  push:\n    # which branches\n    branches:\n      - main\n" +
				"jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n",
			want: "# which branches\n    branches",
		},
		{
			name: "closing an event",
			yaml: "on:\n  push:\n    branches:\n      - main\n    # nothing else\n" +
				"jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n",
			want: "# nothing else\n  }",
		},
		{
			name: "above an env entry",
			yaml: commentWorkflowYAML("env:\n  # what it holds\n  FOO: bar\n", ""),
			want: "# what it holds\n  env {",
		},
		{
			name: "trailing an env entry",
			yaml: commentWorkflowYAML("env:\n  FOO: bar # the value\n", ""),
			want: "\"bar\" # the value",
		},
		{
			name: "closing an env mapping",
			yaml: commentWorkflowYAML("env:\n  FOO: bar\n  # and nothing else\n", ""),
			want: "# and nothing else",
		},
		{
			name: "closing a runs-on list",
			yaml: "on:\n  push: {}\njobs:\n  build:\n    runs-on:\n      - ubuntu-latest\n" +
				"      # nothing fancier\n    steps:\n      - run: make\n",
			want: "# nothing fancier\n  }",
		},
		{
			name: "inside a runs-on mapping",
			yaml: "on:\n  push: {}\njobs:\n  build:\n    runs-on:\n      group: big # the fast ones\n" +
				"    steps:\n      - run: make\n",
			want: "\"big\" # the fast ones",
		},
		{
			name: "above a labelled trigger block",
			yaml: "on:\n  workflow_call:\n    # what it takes\n    inputs:\n      env:\n        type: string\n" +
				"jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n",
			want: "# what it takes",
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

// A job's env is written by the same helper as the workflow's, through a
// different caller, so it needs its own case: the tree it reads is one level
// further down.
func TestJobEnvCommentsSurviveUnparse(t *testing.T) {
	hcl := unparse(t, commentWorkflowYAML("", "    env:\n      # inner\n      A: b\n"))

	if !strings.Contains(hcl, "# inner") {
		t.Errorf("want the comment inside the job env, got:\n%s", hcl)
	}
}
