// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A comment above a block was kept for job blocks and dropped for every other
// kind, so a note above an "env", a "runs_on" or an "on" block never reached
// the YAML.
func TestBlockHeadCommentsSurviveParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		job  string
		wf   string
		want string
	}{
		{
			name: "above an on block",
			wf:   "  # when it fires\n  on \"push\" {}\n",
			want: "on:\n  # when it fires\n  push:",
		},
		{
			name: "above a workflow permissions block",
			wf:   "  on \"push\" {}\n  # who can do what\n  permissions {\n    contents = \"read\"\n  }\n",
			want: "# who can do what\npermissions:",
		},
		{
			name: "above a runs_on block",
			job:  "  # where it runs\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n",
			want: "    # where it runs\n    runs-on: ubuntu-latest",
		},
		{
			name: "above an env block",
			job:  "  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  # what the job needs\n  env {\n    name  = \"FOO\"\n    value = \"bar\"\n  }\n",
			want: "    env:\n      # what the job needs\n      FOO: bar",
		},
		{
			name: "above a job container block",
			job:  "  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  # where it runs inside\n  container {\n    image = \"node:20\"\n  }\n",
			want: "    # where it runs inside\n    container:",
		},
		{
			// A blank line breaks the run, the same way it reads on the page.
			name: "a blank line above the block",
			job:  "  # not about it\n\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n",
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			job := tc.job

			if job == "" {
				job = "  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n"
			}

			wf := tc.wf

			if wf == "" {
				wf = "  on \"push\" {}\n"
			}

			out := parseWorkflow(t, blockCommentHCL(job, wf))

			content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
			if err != nil {
				t.Fatal(err)
			}

			yaml := string(content)

			if tc.want == "" {
				if strings.Contains(yaml, "# not about it") {
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

// The block comment has to survive the other direction too, or it is lost on
// whichever hop does not carry it.
func TestBlockHeadCommentsSurviveUnparse(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "above a permissions block",
			yaml: commentWorkflowYAML("# who can do what\npermissions:\n  contents: read\n", ""),
			want: "# who can do what\n  permissions {",
		},
		{
			name: "above a job env block",
			yaml: commentWorkflowYAML("", "    # what the job needs\n    env:\n      FOO: bar\n"),
			want: "# what the job needs\n  env {",
		},
		{
			name: "above a runs-on key",
			yaml: "on:\n  push: {}\njobs:\n  build:\n    # where it runs\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n",
			want: "# where it runs\n  runs_on {",
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

// The generation markers are cinzel's own note at the top of every file it
// writes, which a YAML reader hands back as the comment above the first key.
// Carrying them into the HCL copied them into the source a person edits, one
// more line on each roundtrip.
func TestGeneratedMarkersAreNotCarriedIntoHCL(t *testing.T) {
	yaml := "# generated-by: cinzel\n# cinzel-provider: github\n" +
		commentWorkflowYAML("name: ci\n", "")

	hcl := unparse(t, yaml)

	for _, gone := range []string{"generated-by", "cinzel-provider"} {
		if strings.Contains(hcl, gone) {
			t.Errorf("the marker %q reached the HCL:\n%s", gone, hcl)
		}
	}
}

// A comment above a real key still reaches the HCL when the markers sit above
// it, which is every generated file.
func TestACommentBelowTheMarkersStillReachesHCL(t *testing.T) {
	yaml := "# generated-by: cinzel\n# cinzel-provider: github\n" +
		commentWorkflowYAML("# what this is for\nname: ci\n", "")

	hcl := unparse(t, yaml)

	if !strings.Contains(hcl, "# what this is for\n  name = \"ci\"") {
		t.Errorf("the comment did not reach the HCL:\n%s", hcl)
	}
}

// blockCommentHCL returns a workflow whose job body and workflow body are the
// given lines, which is where each case writes its commented block.
func blockCommentHCL(job, wf string) string {
	return "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n" + job + "  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n" + wf + "  jobs = [job.b]\n}\n"
}
