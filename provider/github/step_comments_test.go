// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Steps decode through a typed path with no source ranges, so no step
// attribute except "uses" carried a comment. A note on a step's name, run or
// if, or on one of its env values, never reached the workflow file.
func TestStepCommentsSurviveParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		step string
		want string
	}{
		{
			name: "above the step block",
			step: "# what this step is for\nstep \"s\" {\n  run = \"make\"\n}\n",
			want: "      # what this step is for\n      - id: s",
		},
		{
			name: "above an attribute",
			step: "step \"s\" {\n  # how it is built\n  run = \"make\"\n}\n",
			want: "        # how it is built\n        run: make",
		},
		{
			name: "beside an attribute",
			step: "step \"s\" {\n  run = \"make\" # the default target\n}\n",
			want: "run: make # the default target",
		},
		{
			name: "above and beside the same attribute",
			step: "step \"s\" {\n  # how it is built\n  run = \"make\" # the default target\n}\n",
			want: "        # how it is built\n        run: make # the default target",
		},
		{
			name: "on a name attribute",
			step: "step \"s\" {\n  name = \"build\" # what it is called\n  run  = \"make\"\n}\n",
			want: "name: build # what it is called",
		},
		{
			name: "on an env value",
			step: "step \"s\" {\n  run = \"make\"\n\n  env {\n    name = \"FOO\"\n    # what it points at\n    value = \"bar\" # the usual one\n  }\n}\n",
			want: "          # what it points at\n          FOO: bar # the usual one",
		},
		{
			// The pin tag is written on the version, and the whole block
			// becomes one "uses" key, so the comment lands beside that key.
			name: "on a uses version",
			step: "step \"s\" {\n  uses {\n    action  = \"actions/checkout\"\n    version = \"abc123\" # v4\n  }\n}\n",
			want: "uses: actions/checkout@abc123 # v4",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := parseWorkflow(t, stepCommentHCL(tc.step))

			content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(content), tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, content)
			}
		})
	}
}

// The other direction: a comment written on a step in YAML has to reach the
// HCL, or it is lost the first time anyone converts a workflow they wrote.
func TestStepCommentsSurviveUnparse(t *testing.T) {
	for _, tc := range []struct {
		name  string
		steps string
		want  string
	}{
		{
			name:  "above the whole step",
			steps: "      # what this step is for\n      - run: make\n",
			want:  "# what this step is for\nstep \"make\" {",
		},
		{
			name:  "above an attribute",
			steps: "      - # how it is built\n        run: make\n",
			want:  "# how it is built\n  run = \"make\"",
		},
		{
			name:  "beside an attribute",
			steps: "      - run: make # the default target\n",
			want:  "run = \"make\" # the default target",
		},
		{
			name:  "on a name attribute",
			steps: "      - name: build # what it is called\n        run: make\n",
			want:  "name = \"build\" # what it is called",
		},
		{
			name:  "on an env value",
			steps: "      - run: make\n        env:\n          FOO: bar # the usual one\n",
			want:  "value = \"bar\" # the usual one",
		},
		{
			// The pin tag rides the version attribute, which is the half of
			// the uses block that carries the SHA.
			name:  "on a uses pin",
			steps: "      - uses: actions/checkout@abc123 # v4\n",
			want:  "version = \"abc123\" # v4",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl := unparse(t, stepCommentYAML(tc.steps))

			if !strings.Contains(hcl, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, hcl)
			}
		})
	}
}

// The pin tag is the only record of which version a SHA stands for, so losing
// it on a conversion loses the one thing the SHA cannot say.
func TestThePinCommentSurvivesARoundtrip(t *testing.T) {
	yaml := stepCommentYAML("      - uses: actions/checkout@abc123 # v4\n")

	hcl := unparse(t, yaml)

	if !strings.Contains(hcl, "version = \"abc123\" # v4") {
		t.Fatalf("the pin tag did not reach the HCL:\n%s", hcl)
	}

	out := parseWorkflow(t, hcl)

	content, err := os.ReadFile(filepath.Join(out, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "uses: actions/checkout@abc123 # v4") {
		t.Errorf("the pin tag did not come back to the YAML:\n%s", content)
	}
}

// Every form at once, to catch a fix that carries one comment by moving
// another.
func TestEveryStepCommentSurvivesARoundtrip(t *testing.T) {
	yaml := stepCommentYAML("      # check the code out\n" +
		"      - uses: actions/checkout@abc123 # v4\n" +
		"      - name: build # the main one\n" +
		"        # how it is built\n" +
		"        run: make\n" +
		"        env:\n          # points at prod\n          NODE_ENV: production\n")

	out := parseWorkflow(t, unparse(t, yaml))

	content, err := os.ReadFile(filepath.Join(out, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"# check the code out",
		"uses: actions/checkout@abc123 # v4",
		"name: build # the main one",
		"# how it is built",
		"# points at prod",
	} {
		if !strings.Contains(string(content), want) {
			t.Errorf("want %q in:\n%s", want, content)
		}
	}
}

// Two jobs running the same action collapse to one step block, so two
// differently-commented uses of it had one block to live on and the second
// comment was dropped. A comment is part of a step's identity.
func TestStepsDifferingOnlyByCommentAreNotMerged(t *testing.T) {
	yaml := "on:\n  push: {}\njobs:\n" +
		"  a:\n    runs-on: ubuntu-latest\n    steps:\n      # checkout for the build\n      - uses: actions/checkout@v4\n" +
		"  b:\n    runs-on: ubuntu-latest\n    steps:\n      # checkout for the tests\n      - uses: actions/checkout@v4\n"

	hcl := unparse(t, yaml)

	for _, want := range []string{"# checkout for the build", "# checkout for the tests"} {
		if !strings.Contains(hcl, want) {
			t.Errorf("want %q in:\n%s", want, hcl)
		}
	}

	if blocks := strings.Count(hcl, "step \""); blocks != 2 {
		t.Errorf("want 2 step blocks, got %d:\n%s", blocks, hcl)
	}
}

// The dedup still has to work. Folding comments into the fingerprint must not
// split two steps that are the same all the way down.
func TestIdenticalStepsStillMerge(t *testing.T) {
	yaml := "on:\n  push: {}\njobs:\n" +
		"  a:\n    runs-on: ubuntu-latest\n    steps:\n      # shared\n      - uses: actions/checkout@v4\n" +
		"  b:\n    runs-on: ubuntu-latest\n    steps:\n      # shared\n      - uses: actions/checkout@v4\n"

	hcl := unparse(t, yaml)

	if blocks := strings.Count(hcl, "step \""); blocks != 1 {
		t.Errorf("want 1 step block, got %d:\n%s", blocks, hcl)
	}
}

// A run script is written as a heredoc, which ends on its closing marker's own
// line. A comment after the tokens would land past that marker rather than
// beside the attribute, so it is written above instead.
func TestACommentOnAHeredocRunIsWrittenAbove(t *testing.T) {
	yaml := stepCommentYAML("      - run: | # what it does\n          make\n          make test\n")

	hcl := unparse(t, yaml)

	if !strings.Contains(hcl, "# what it does\n  run = <<-EOF") {
		t.Errorf("the comment was not written above the heredoc:\n%s", hcl)
	}

	// It has to still be valid HCL, which is the thing a misplaced comment
	// breaks.
	parseWorkflow(t, hcl)
}

func stepCommentHCL(step string) string {
	label := strings.Split(strings.Split(step, "step \"")[1], "\"")[0]

	return step + "\njob \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step." + label + "]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n"
}

func stepCommentYAML(steps string) string {
	return "on:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n" + steps
}
