// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Unparse wrote no attribute comment at all: the decoded document cannot hold
// one, and only the comment above a job was read off the node tree beside it.
// Everything else written in a workflow file was dropped on the way to HCL.
func TestAttributeCommentsSurviveUnparse(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want []string
		gone []string
	}{
		{
			name: "above a workflow attribute",
			yaml: commentWorkflowYAML("# what this is for\nname: ci\n", ""),
			want: []string{"# what this is for\n  name = \"ci\""},
		},
		{
			name: "beside a workflow attribute",
			yaml: commentWorkflowYAML("name: ci # what this is for\n", ""),
			want: []string{"name = \"ci\" # what this is for"},
		},
		{
			name: "above an attribute in a nested block",
			yaml: commentWorkflowYAML("permissions:\n  # we only need to read\n  contents: read\n", ""),
			want: []string{"# we only need to read\n    contents = \"read\""},
		},
		{
			name: "beside an attribute in a nested block",
			yaml: commentWorkflowYAML("permissions:\n  contents: read # all we need\n", ""),
			want: []string{"contents = \"read\" # all we need"},
		},
		{
			name: "both forms on the same attribute",
			yaml: commentWorkflowYAML("permissions:\n  # above\n  contents: read # beside\n", ""),
			want: []string{"# above\n    contents = \"read\" # beside"},
		},
		{
			name: "above a job attribute",
			yaml: commentWorkflowYAML("", "    # no longer than that\n    timeout-minutes: 5\n"),
			want: []string{"# no longer than that\n  timeout_minutes = 5"},
		},
		{
			name: "beside a job attribute",
			yaml: commentWorkflowYAML("", "    timeout-minutes: 5 # no longer than that\n"),
			want: []string{"timeout_minutes = 5 # no longer than that"},
		},
		{
			name: "an attribute with no comment is untouched",
			yaml: commentWorkflowYAML("name: ci\n", ""),
			gone: []string{"#"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl := unparse(t, tc.yaml)

			for _, want := range tc.want {
				if !strings.Contains(hcl, want) {
					t.Errorf("want %q in:\n%s", want, hcl)
				}
			}

			for _, gone := range tc.gone {
				if strings.Contains(hcl, gone) {
					t.Errorf("want %q absent from:\n%s", gone, hcl)
				}
			}
		})
	}
}

// A comment is prose. Its spacing and its "#" count are the author's, so the
// text reaches the HCL as it was written.
func TestAttributeCommentIsVerbatimOnUnparse(t *testing.T) {
	yaml := commentWorkflowYAML("permissions:\n  contents: read ##no space and   wide   spacing\n", "")

	hcl := unparse(t, yaml)

	if !strings.Contains(hcl, "##no space and   wide   spacing") {
		t.Errorf("the comment was rewritten on the way out:\n%s", hcl)
	}
}

// A comment written in YAML has to come back to YAML unchanged, or it is lost
// on whichever hop does not carry it.
func TestAttributeCommentsSurviveARoundtrip(t *testing.T) {
	yaml := commentWorkflowYAML(
		"permissions:\n  # we only need to read\n  contents: read\n  issues: write # and file bugs\n",
		"    timeout-minutes: 5 # no longer than that\n")

	out := parseWorkflow(t, unparse(t, yaml))

	content, err := os.ReadFile(filepath.Join(out, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"  # we only need to read\n  contents: read",
		"issues: write # and file bugs",
		"timeout-minutes: 5 # no longer than that",
	} {
		if !strings.Contains(string(content), want) {
			t.Errorf("want %q in:\n%s", want, content)
		}
	}
}

// commentWorkflowYAML returns a workflow carrying the given top-level lines and
// the given extra lines inside its one job.
func commentWorkflowYAML(top, jobLines string) string {
	return "on:\n  push: {}\n" + top + "jobs:\n  build:\n    runs-on: ubuntu-latest\n" +
		jobLines + "    steps:\n      - run: make\n"
}
