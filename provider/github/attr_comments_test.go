// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A comment above an attribute is a different thing from one trailing it: it is
// usually a sentence about the line below, not a note on the value. Parse read
// only the trailing form, so the one written above was dropped.
func TestAttributeHeadCommentsSurviveParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want []string
		gone []string
	}{
		{
			name: "one line above the attribute",
			body: "  permissions {\n    # we only need to read\n    contents = \"read\"\n  }\n",
			want: []string{"  # we only need to read\n  contents: read"},
		},
		{
			name: "two lines above the attribute",
			body: "  permissions {\n    # first\n    # second\n    contents = \"read\"\n  }\n",
			want: []string{"  # first\n  # second\n  contents: read"},
		},
		{
			// The two forms are independent and both land where written.
			name: "head and trailing on the same attribute",
			body: "  permissions {\n    # above\n    contents = \"read\" # beside\n  }\n",
			want: []string{"  # above\n  contents: read # beside"},
		},
		{
			// A blank line breaks the run, the same way it reads on the page.
			name: "a blank line between the comment and the attribute",
			body: "  permissions {\n    # not about it\n\n    contents = \"read\"\n  }\n",
			gone: []string{"# not about it"},
		},
		{
			// It belongs to what it trails, so taking it as a head comment
			// would move it somewhere it was not written.
			name: "a comment trailing the attribute above",
			body: "  permissions {\n    contents = \"read\" # about contents\n    issues   = \"write\"\n  }\n",
			want: []string{"permissions:\n  contents: read # about contents\n  issues: write\n"},
		},
		{
			name: "only the commented attribute carries one",
			body: "  permissions {\n    # about contents\n    contents = \"read\"\n    issues   = \"write\"\n  }\n",
			want: []string{"  # about contents\n  contents: read"},
			gone: []string{"# about contents\n  issues"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := parseWorkflow(t, attrCommentHCL(tc.body))

			content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
			if err != nil {
				t.Fatal(err)
			}

			yaml := string(content)

			for _, want := range tc.want {
				if !strings.Contains(yaml, want) {
					t.Errorf("want %q in:\n%s", want, yaml)
				}
			}

			for _, gone := range tc.gone {
				if strings.Contains(yaml, gone) {
					t.Errorf("want %q absent from:\n%s", gone, yaml)
				}
			}
		})
	}
}

// A comment is prose. Its spacing and its "#" count are the author's, so the
// text reaches the YAML as it was written.
func TestAttributeHeadCommentIsVerbatim(t *testing.T) {
	body := "  permissions {\n    ##no space and   wide   spacing\n    contents = \"read\"\n  }\n"

	out := parseWorkflow(t, attrCommentHCL(body))

	content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "##no space and   wide   spacing") {
		t.Errorf("the comment was rewritten on the way out:\n%s", content)
	}
}

// Attributes inside a job body walk the same path, so the comment above one is
// read there too.
func TestAJobAttributeHeadCommentSurvivesParse(t *testing.T) {
	hcl := "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		"  # only when it matters\n  if = \"true\"\n  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n  jobs = [job.b]\n}\n"

	out := parseWorkflow(t, hcl)

	content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "# only when it matters\n    if:") {
		t.Errorf("the comment did not reach the YAML:\n%s", content)
	}
}

// attrCommentHCL returns a workflow whose body is the given lines, which is
// where each case writes its commented attribute.
func attrCommentHCL(body string) string {
	return "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n  on \"push\" {}\n" + body + "  jobs = [job.b]\n}\n"
}

// The foot comment has a roundtrip of its own in foot_comments_test.go. This is
// the other two positions, starting from the HCL a user writes rather than from
// YAML, so a comment has to survive both hops back to where it began.
func TestHeadAndTrailingCommentsSurviveARoundtripFromHCL(t *testing.T) {
	const src = "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"# why this job is here\n" +
		"job \"b\" {\n" +
		"  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n" +
		"  # no longer than that\n" +
		"  timeout_minutes = 5 # five\n\n" +
		"  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n\n  on \"push\" {}\n\n  jobs = [job.b]\n}\n"

	out := parseWorkflow(t, src)

	content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	hcl := unparse(t, string(content))

	for _, want := range []string{
		"# why this job is here",
		"# no longer than that",
		"# five",
	} {
		if !strings.Contains(hcl, want) {
			t.Errorf("want %q after a roundtrip, got:\n%s", want, hcl)
		}
	}
}
