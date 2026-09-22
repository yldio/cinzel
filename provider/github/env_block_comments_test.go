// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An env or with entry is written as a block, so there are three places a
// comment sits on one: above the block, above the value inside it, and at the
// end of the body. Each path read a different subset. The step path read the
// value and nothing else, so a note above an "env {" was dropped; the job path
// read the block and nothing else, so a note above its "value =" was dropped.
// Both are the author's, and both belong above the one YAML key the block
// becomes.
func TestEveryEnvBlockCommentSurvivesParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want []string
	}{
		{
			name: "above a step env block",
			hcl: stepWithBody("  run = \"echo\"\n\n" +
				"  # what it holds\n  env {\n    name  = \"FOO\"\n    value = \"bar\"\n  }\n"),
			want: []string{"# what it holds\n          FOO: bar"},
		},
		{
			name: "above a step with block",
			hcl: stepWithBody("  uses {\n    action  = \"actions/checkout\"\n    version = \"v4\"\n  }\n\n" +
				"  # how deep\n  with {\n    name  = \"fetch-depth\"\n    value = \"0\"\n  }\n"),
			want: []string{"# how deep\n          fetch-depth:"},
		},
		{
			name: "closing a step env block",
			hcl: stepWithBody("  run = \"echo\"\n\n" +
				"  env {\n    name  = \"FOO\"\n    value = \"bar\"\n    # and nothing else\n  }\n"),
			want: []string{"# and nothing else"},
		},
		{
			name: "above a step env block and above its value",
			hcl: stepWithBody("  run = \"echo\"\n\n" +
				"  # what it holds\n  env {\n    name = \"FOO\"\n\n    # the usual one\n    value = \"bar\"\n  }\n"),
			want: []string{"# what it holds\n          # the usual one\n          FOO: bar"},
		},
		{
			name: "above a job env value",
			hcl: jobWithBody("  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n" +
				"  env {\n    name = \"FOO\"\n\n    # the usual one\n    value = \"bar\"\n  }\n"),
			want: []string{"# the usual one\n      FOO: bar"},
		},
		{
			name: "beside a job env value",
			hcl: jobWithBody("  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n" +
				"  env {\n    name  = \"FOO\"\n    value = \"bar\" # the usual one\n  }\n"),
			want: []string{"FOO: bar # the usual one"},
		},
		{
			name: "above a job env block and above its value",
			hcl: jobWithBody("  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n" +
				"  # what it holds\n  env {\n    name = \"FOO\"\n\n    # the usual one\n    value = \"bar\"\n  }\n"),
			want: []string{"# what it holds\n      # the usual one\n      FOO: bar"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := readWorkflow(t, parseWorkflow(t, tc.hcl))

			for _, want := range tc.want {
				if !strings.Contains(content, want) {
					t.Errorf("want %q in:\n%s", want, content)
				}
			}
		})
	}
}

// The comment closing an env or with mapping is written above the key that
// comes last, which is the block it ends up inside and so where the next parse
// reads it from. Written after the blocks instead, it landed in the
// surrounding body where nothing read it: the comment survived one conversion
// and was gone by the second.
func TestAClosingEnvCommentIsWrittenInsideTheLastBlock(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "closing a job env",
			yaml: commentWorkflowYAML("", "    env:\n      A: one\n      B: two\n      # and nothing else\n"),
			want: "value = \"two\"\n    # and nothing else\n  }",
		},
		{
			name: "closing a step env",
			yaml: "on:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n" +
				"      - run: make\n        env:\n          A: one\n          B: two\n          # and nothing else\n",
			want: "value = \"two\"\n    # and nothing else\n  }",
		},
		{
			name: "closing a workflow env",
			yaml: commentWorkflowYAML("env:\n  A: one\n  B: two\n  # and nothing else\n", ""),
			want: "value = \"two\"\n    # and nothing else\n  }",
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

// The whole point of reading both positions: what the HCL says has to still be
// there after a trip through YAML and back, or the conversion eats a comment
// every time it is run.
func TestEnvBlockCommentsAreStableAcrossTwoPasses(t *testing.T) {
	hcl := stepWithBody("  run = \"echo\"\n\n" +
		"  # what it holds\n  env {\n    name = \"FOO\"\n\n" +
		"    # the usual one\n    value = \"bar\" # inline\n    # and nothing else\n  }\n")

	first := readWorkflow(t, parseWorkflow(t, hcl))
	second := readWorkflow(t, parseWorkflow(t, unparse(t, first)))

	if first != second {
		t.Errorf("the YAML changed on the second pass\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	for _, want := range []string{"# what it holds", "# the usual one", "# inline", "# and nothing else"} {
		if !strings.Contains(first, want) {
			t.Errorf("a comment was lost: %q\nYAML:\n%s", want, first)
		}
	}
}

// readWorkflow returns the single workflow file written into dir. Named for
// the workflow rather than the file it came from, so the second pass writes it
// under a different name than the first.
func readWorkflow(t *testing.T, dir string) string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("want one workflow in %s, got %d", dir, len(entries))
	}

	content, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// stepWithBody returns a workflow whose one step has the given body.
func stepWithBody(body string) string {
	return "step \"s\" {\n" + body + "}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.s]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n"
}

// jobWithBody returns a workflow whose one job has the given body.
func jobWithBody(body string) string {
	return "step \"s\" {\n  run = \"echo\"\n}\n\n" +
		"job \"b\" {\n" + body + "  steps = [step.s]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n"
}
