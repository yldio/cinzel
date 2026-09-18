// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A comment written at the end of a block with nothing after it was read by
// neither direction, so a note closing a job, a step or a workflow was dropped.
func TestFootCommentsSurviveParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		step string
		job  string
		wf   string
		want string
	}{
		{
			name: "closing a step",
			step: "  run = \"echo hi\"\n  # that is all it does\n",
			want: "        run: echo hi\n        # that is all it does",
		},
		{
			name: "closing a job",
			job:  "  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.hi]\n  # that is the whole job\n",
			want: "    # that is the whole job",
		},
		{
			name: "closing a runs_on block",
			job:  "  runs_on {\n    runners = \"ubuntu-latest\"\n    # nothing fancier\n  }\n  steps = [step.hi]\n",
			want: "    runs-on: ubuntu-latest\n    # nothing fancier",
		},
		{
			name: "closing an env block",
			job:  "  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  env {\n    name  = \"FOO\"\n    value = \"bar\"\n    # and nothing else\n  }\n  steps = [step.hi]\n",
			want: "      FOO: bar\n      # and nothing else",
		},
		{
			name: "closing an on block",
			wf:   "  on \"push\" {\n    branches = [\"main\"]\n    # only that branch\n  }\n  jobs = [job.b]\n",
			want: "      - main\n    # only that branch",
		},
		{
			name: "closing a workflow",
			wf:   "  on \"push\" {}\n  jobs = [job.b]\n  # that is the whole workflow\n",
			want: "# that is the whole workflow",
		},
		{
			// A blank line breaks the run, the same way it reads on the page.
			// The comment no longer sits against the closing brace, so it is
			// no longer what closes the block.
			name: "a blank line below the comment",
			step: "  run = \"echo hi\"\n  # not about the step\n\n",
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := parseWorkflow(t, footCommentHCL(tc.step, tc.job, tc.wf))

			content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
			if err != nil {
				t.Fatal(err)
			}

			yaml := string(content)

			if tc.want == "" {
				if strings.Contains(yaml, "# not about the step") {
					t.Errorf("a comment broken off by a blank line was taken anyway:\n%s", yaml)
				}

				return
			}

			if !strings.Contains(yaml, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, yaml)
			}
		})
	}
}

// The foot comment has to survive the other direction too, or it is lost on
// whichever hop does not carry it.
func TestFootCommentsSurviveUnparse(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "closing a step",
			yaml: "on:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n        # that is all it does\n",
			want: "# that is all it does\n}",
		},
		{
			name: "closing a job",
			yaml: "on:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n    # that is the whole job\n",
			want: "# that is the whole job\n}",
		},
		{
			name: "closing a permissions block",
			yaml: commentWorkflowYAML("permissions:\n  contents: read\n  # read is enough\n", ""),
			want: "# read is enough\n  }",
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

// A foot comment belongs to the block it closes, so it has to come back
// closing the same block rather than attached to the last thing inside it.
func TestAFootCommentSurvivesARoundtrip(t *testing.T) {
	const src = "step \"hi\" {\n  run = \"echo hi\"\n  # that is all it does\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n" +
		"  steps = [step.hi]\n  # that is the whole job\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n\n  on \"push\" {}\n\n  jobs = [job.b]\n}\n"

	out := parseWorkflow(t, src)

	content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	hcl := unparse(t, string(content))

	for _, want := range []string{
		"# that is all it does\n}",
		"# that is the whole job\n}",
	} {
		if !strings.Contains(hcl, want) {
			t.Errorf("want %q after a roundtrip, got:\n%s", want, hcl)
		}
	}
}

// footCommentHCL returns a workflow whose step, job and workflow bodies are the
// given lines, which is where each case writes its closing comment.
func footCommentHCL(step, job, wf string) string {
	if step == "" {
		step = "  run = \"echo hi\"\n"
	}

	if job == "" {
		job = "  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.hi]\n"
	}

	if wf == "" {
		wf = "  on \"push\" {}\n  jobs = [job.b]\n"
	}

	return "step \"hi\" {\n" + step + "}\n\n" +
		"job \"b\" {\n" + job + "}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n" + wf + "}\n"
}
