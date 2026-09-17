// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The comment above a job block is the one thing the attributes below it
// cannot say. Parse dropped it, so a note written in HCL never reached the
// workflow file anyone reads on GitHub.
func TestJobCommentsSurviveParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		jobs string
		want []string
		gone []string
	}{
		{
			name: "one line above the block",
			jobs: "# only runs on main\n" + commentJobHCL("check", ""),
			want: []string{"  # only runs on main\n  check:"},
		},
		{
			name: "two lines above the block",
			jobs: "# first\n# second\n" + commentJobHCL("check", ""),
			want: []string{"  # first\n  # second\n  check:"},
		},
		{
			name: "a job with no comment is untouched",
			jobs: commentJobHCL("check", ""),
			gone: []string{"#"},
		},
		{
			// The comment is written above the block, so it is found by the
			// label; the key it lands under comes from the id attribute.
			name: "a job whose key is not its label",
			jobs: "# about the renamed one\n" + commentJobHCL("check", `  id = "check-something"`),
			want: []string{"  # about the renamed one\n  check-something:"},
		},
		{
			// A blank line breaks the run, the same way it reads on the page.
			name: "a blank line between the comment and the block",
			jobs: "# not about the job\n\n" + commentJobHCL("check", ""),
			gone: []string{"# not about the job"},
		},
		{
			// It belongs to what it trails. Taking it would move it somewhere
			// it was not written.
			name: "a comment trailing code is left where it is",
			jobs: commentJobHCL("check", "") + "# above nothing\n",
			gone: []string{"  # above nothing\n  check:"},
		},
		{
			name: "only the commented job carries one",
			jobs: "# about first\n" + commentJobHCL("first", "") + commentJobHCL("second", ""),
			want: []string{"  # about first\n  first:"},
			gone: []string{"# about first\n  second:"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := parseWorkflow(t, commentWorkflowHCL(tc.jobs))

			content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
			if err != nil {
				t.Fatal(err)
			}

			yaml := string(content)

			// Every generated file opens with the marker comment, which is
			// not what any of these cases is about.
			body := yaml[strings.Index(yaml, "on:"):]

			for _, want := range tc.want {
				if !strings.Contains(body, want) {
					t.Errorf("want %q in:\n%s", want, body)
				}
			}

			for _, gone := range tc.gone {
				if strings.Contains(body, gone) {
					t.Errorf("want %q absent from:\n%s", gone, body)
				}
			}
		})
	}
}

// The comment has to survive both directions, or it is lost on whichever hop
// does not carry it.
func TestAJobCommentSurvivesARoundtrip(t *testing.T) {
	yaml := "name: ci\non:\n  push: {}\njobs:\n  # why this job is here\n" +
		"  check-something:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"

	hcl := unparse(t, yaml)

	if !strings.Contains(hcl, "# why this job is here\njob \"check_something\"") {
		t.Fatalf("the comment did not reach the HCL:\n%s", hcl)
	}

	out := parseWorkflow(t, hcl)

	// unparse writes the YAML as workflow.yaml, so that is the filename the
	// generated workflow block carries back.
	content, err := os.ReadFile(filepath.Join(out, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "  # why this job is here\n  check-something:") {
		t.Errorf("the comment did not come back to the YAML:\n%s", content)
	}
}

func commentJobHCL(label, extra string) string {
	block := "job \"" + label + "\" {\n"

	if extra != "" {
		block += extra + "\n"
	}

	return block + "  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.hi]\n}\n\n"
}

func commentWorkflowHCL(jobs string) string {
	refs := []string{}

	for _, line := range strings.Split(jobs, "\n") {
		if strings.HasPrefix(line, "job \"") {
			refs = append(refs, "job."+strings.Split(line, "\"")[1])
		}
	}

	return "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" + jobs +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [" +
		strings.Join(refs, ", ") + "]\n}\n"
}
