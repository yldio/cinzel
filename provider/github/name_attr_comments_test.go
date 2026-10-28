// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A name/value block becomes one YAML key, and the comment on its "name =" is
// a comment on that key. Only the value's line was read, so a note written
// beside the name was dropped: the parse exited 0 and the generated file came
// out byte for byte what it was before, so "diff" reported nothing either and
// the note had no way of showing it had gone.
func TestANameAttributeCommentReachesTheKeyItNames(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want string
	}{
		{
			name: "a job env block",
			hcl: namedBlockHCL("job", "  env {\n"+
				"    name  = \"E\" # why this is set\n"+
				"    value = \"v\"\n  }\n"),
			want: "E: v # why this is set",
		},
		{
			name: "a job output block",
			hcl: namedBlockHCL("job", "  output {\n"+
				"    name  = \"O\" # what reads it\n"+
				"    value = \"v\"\n  }\n"),
			want: "O: v # what reads it",
		},
		{
			name: "a workflow env block",
			hcl: namedBlockHCL("workflow", "  env {\n"+
				"    name  = \"E\" # why this is set\n"+
				"    value = \"v\"\n  }\n"),
			want: "E: v # why this is set",
		},
		{
			name: "a step env block",
			hcl: namedBlockHCL("step", "  env {\n"+
				"    name  = \"E\" # why this is set\n"+
				"    value = \"v\"\n  }\n"),
			want: "E: v # why this is set",
		},
		{
			name: "a step with block",
			hcl: namedBlockHCL("step", "  with {\n"+
				"    name  = \"W\" # what this input does\n"+
				"    value = \"v\"\n  }\n"),
			want: "W: v # what this input does",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := parsedYAML(t, tc.hcl); !strings.Contains(got, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, got)
			}
		})
	}
}

// A run of comments above the "name =" belongs above the key the block becomes,
// the same as one written above the block or above its value.
func TestANameAttributeHeadCommentReachesTheKeyItNames(t *testing.T) {
	hcl := namedBlockHCL("job", "  env {\n"+
		"    # why this is set\n"+
		"    name  = \"E\"\n"+
		"    value = \"v\"\n  }\n")

	want := "# why this is set\n      E: v"

	if got := parsedYAML(t, hcl); !strings.Contains(got, want) {
		t.Errorf("want %q in:\n%s", want, got)
	}
}

// The key has one line, and it belongs to the value: that is the attribute it
// carries, and a pin tag or a unit lands there. A comment beside the name has
// nowhere else to sit, so it stacks above the key rather than taking a line
// that is already spoken for.
func TestACommentOnBothNameAndValueKeepsBoth(t *testing.T) {
	hcl := namedBlockHCL("job", "  env {\n"+
		"    name  = \"E\" # about the name\n"+
		"    value = \"v\" # about the value\n  }\n")

	yaml := parsedYAML(t, hcl)

	for _, want := range []string{"# about the name", "E: v # about the value"} {
		if !strings.Contains(yaml, want) {
			t.Errorf("want %q in:\n%s", want, yaml)
		}
	}
}

// Every comment is prose, so two of them above one key read as one run rather
// than one of them being chosen over the other.
func TestCommentsAboveTheBlockAndTheNameAreJoined(t *testing.T) {
	hcl := namedBlockHCL("job", "  # above the block\n  env {\n"+
		"    # above the name\n"+
		"    name  = \"E\"\n"+
		"    value = \"v\"\n  }\n")

	want := "# above the block\n      # above the name\n      E: v"

	if got := parsedYAML(t, hcl); !strings.Contains(got, want) {
		t.Errorf("want %q in:\n%s", want, got)
	}
}

// namedBlockHCL returns a workflow carrying body in the named scope, which is
// where each case writes its name/value block.
func namedBlockHCL(scope, body string) string {
	step, job, workflow := "", "", ""

	switch scope {
	case "step":
		step = body
	case "job":
		job = body
	case "workflow":
		workflow = body
	}

	return "step \"hi\" {\n  run = \"echo hi\"\n" + step + "}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		job + "  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  name = \"W\"\n  on \"push\" {}\n" +
		workflow + "  jobs = [job.b]\n}\n"
}
