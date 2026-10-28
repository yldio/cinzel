// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A comment written where a type assertion reads the value stopped the parse
// outright: the wrapper carrying it is not the type the reader expected, and
// the failure was reported as the value being wrong rather than as a comment
// being in the way. A note is prose, so it cannot decide whether a workflow
// converts at all.
func TestACommentDoesNotRefuseTheValueItSitsBeside(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want string
	}{
		{
			// "strategy.matrix.variable entries require a non-empty name"
			name: "beside a matrix variable name",
			hcl: matrixHCL("      variable {\n" +
				"        name  = \"os\" # the runners we support\n" +
				"        value = [\"ubuntu-latest\"]\n      }\n"),
			want: "os: # the runners we support",
		},
		{
			// "strategy.matrix.variable must be an object or list"
			name: "above a matrix variable block",
			hcl: matrixHCL("      # the runners we support\n" +
				"      variable {\n" +
				"        name  = \"os\"\n" +
				"        value = [\"ubuntu-latest\"]\n      }\n"),
			want: "# the runners we support\n        os:",
		},
		{
			// "schedule cron must be a string or list"
			name: "beside a schedule cron",
			hcl:  scheduleHCL("    cron = [\"0 0 * * *\"] # nightly\n"),
			want: "# nightly",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := parsedYAML(t, tc.hcl)

			if !strings.Contains(got, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, got)
			}
		})
	}
}

// A matrix axis written as a plain attribute and one written as a variable
// block are the same axis, so a comment on either reaches the YAML. The value
// is the line the axis is emitted on, so its comment is the one that trails.
func TestAMatrixAxisKeepsTheCommentOnItsValue(t *testing.T) {
	hcl := matrixHCL("      variable {\n" +
		"        name  = \"os\"\n" +
		"        value = \"ubuntu-latest\" # only one for now\n      }\n")

	want := "os: ubuntu-latest # only one for now"

	if got := parsedYAML(t, hcl); !strings.Contains(got, want) {
		t.Errorf("want %q in:\n%s", want, got)
	}
}

// "jobs" and "steps" are reference lists rather than values, so the key each
// one becomes is the only place a comment on that line can go.
func TestAReferenceListKeepsItsComment(t *testing.T) {
	hcl := "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		"  # what it does\n  steps = [step.hi] # in order\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  name = \"W\"\n  on \"push\" {}\n" +
		"  # what this builds\n  jobs = [job.b] # just the one\n}\n"

	yaml := parsedYAML(t, hcl)

	for _, want := range []string{
		"# what this builds\njobs: # just the one",
		"# what it does\n    steps: # in order",
	} {
		if !strings.Contains(yaml, want) {
			t.Errorf("want %q in:\n%s", want, yaml)
		}
	}
}

// A uses block becomes one "uses" key with one line. The version's comment has
// first claim on it, since that is where a pin tag lands, but an action written
// without a version has no reason to lose the comment beside it.
func TestAUsesBlockKeepsACommentOnEitherAttribute(t *testing.T) {
	for _, tc := range []struct {
		name string
		uses string
		want string
	}{
		{
			name: "on the version",
			uses: "  uses {\n    action  = \"actions/checkout\"\n" +
				"    version = \"v4\" # pinned\n  }\n",
			want: "uses: actions/checkout@v4 # pinned",
		},
		{
			name: "on the action, with no version comment",
			uses: "  uses {\n    action  = \"actions/checkout\" # the official one\n" +
				"    version = \"v4\"\n  }\n",
			want: "uses: actions/checkout@v4 # the official one",
		},
		{
			// The version is where a pin tag is written, so it wins the line.
			name: "on both",
			uses: "  uses {\n    action  = \"actions/checkout\" # the official one\n" +
				"    version = \"v4\" # pinned\n  }\n",
			want: "uses: actions/checkout@v4 # pinned",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl := "step \"hi\" {\n" + tc.uses + "}\n\n" +
				"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
				"  steps = [step.hi]\n}\n\n" +
				"workflow \"w\" {\n  filename = \"w\"\n  name = \"W\"\n" +
				"  on \"push\" {}\n  jobs = [job.b]\n}\n"

			if got := parsedYAML(t, hcl); !strings.Contains(got, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, got)
			}
		})
	}
}

// matrixHCL returns a workflow whose matrix holds the given block.
func matrixHCL(variable string) string {
	return "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		"  strategy {\n    matrix {\n" + variable + "    }\n  }\n" +
		"  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  name = \"W\"\n" +
		"  on \"push\" {}\n  jobs = [job.b]\n}\n"
}

// scheduleHCL returns a workflow whose schedule event holds the given body.
func scheduleHCL(body string) string {
	return "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		"  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  name = \"W\"\n" +
		"  on \"schedule\" {\n" + body + "  }\n  jobs = [job.b]\n}\n"
}
