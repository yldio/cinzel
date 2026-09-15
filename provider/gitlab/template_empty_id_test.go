// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A hidden key is a template, and its name is what is left after the dot. A
// key of "." leaves nothing, so the template was written with id = "" — which
// the parse direction refuses with "'id' must be a non-empty string". The file
// was written and the command exited 0, so the pipeline could not come back.
//
// The job loop already refuses its own unnamed case. This is the same refusal
// one loop up.
func TestATemplateWithNothingAfterTheDotIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			name: "key is only a dot",
			yaml: `".":
  script: [make dot]
test:
  script: [make test]
`,
		},
		{
			name: "extended by a job",
			yaml: `".":
  script: [make dot]
test:
  extends: ["."]
  script: [make test]
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := unparseGitLab(t, tc.yaml)

			if err == nil {
				t.Fatalf("Unparse() error = nil, want an error\nHCL written:\n%s", got)
			}

			if !strings.Contains(err.Error(), "non-empty string") {
				t.Errorf("Unparse() error = %q, want it to name the empty id", err)
			}
		})
	}
}

// The refusal has to stop at the template with no name left. A template whose
// key carries something after the dot is written as before, and a job extending
// it resolves through the label that template was assigned rather than through
// its sanitized name: two keys sanitizing alike take "base_one" and
// "base_one_2", and a reference built from the name alone would reach the first
// one for both.
func TestATemplateExtendsFollowsTheLabelItWasGiven(t *testing.T) {
	got, err := unparseGitLab(t, `.base-one:
  script: [make a]
.base.one:
  script: [make b]
test:
  extends: [".base.one"]
  script: [make test]
`)
	if err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	for _, want := range []string{
		`template "base_one" {`,
		`template "base_one_2" {`,
		"template.base_one_2,",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HCL missing %q:\n%s", want, got)
		}
	}
}
