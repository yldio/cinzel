// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A "spec" header sits in a document of its own, ahead of the rest of the
// configuration. Unparse used to read only the first document, so everything
// after the "---" was dropped without a word and the header itself was
// mistaken for a job.
func TestSpecHeaderSurvives(t *testing.T) {
	hcl, back := roundtripYAML(t,
		"spec:\n  inputs:\n    env:\n      default: prod\n---\njob1:\n  script:\n    - make\n")

	for _, want := range []string{"spec:", "inputs:", "env:", "default: prod", "---", "job1:", "make"} {
		if !strings.Contains(back, want) {
			t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", want, hcl, back)
		}
	}

	if strings.Index(back, "spec:") > strings.Index(back, "---") {
		t.Errorf("spec must come before the separator\nYAML:\n%s", back)
	}
}

// Every document of a multi-document pipeline is read, not just the first.
func TestSpecHeaderKeepsEveryJob(t *testing.T) {
	hcl, back := roundtripYAML(t,
		"spec:\n  inputs:\n    env:\n      default: prod\n---\nstages:\n  - build\njob1:\n  stage: build\n  script:\n    - make\njob2:\n  stage: build\n  script:\n    - test\n")

	for _, want := range []string{"job1:", "job2:", "stages:", "- build"} {
		if !strings.Contains(back, want) {
			t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", want, hcl, back)
		}
	}
}

// A spec carries the other three keywords GitLab allows it, not just inputs.
func TestSpecHeaderKeepsEveryKeyword(t *testing.T) {
	hcl, back := roundtripYAML(t,
		"spec:\n  description: a component\n  component:\n    - name\n  include:\n    - local: inputs.yml\n  inputs:\n    env: {}\n---\njob1:\n  script:\n    - make\n")

	for _, want := range []string{"description: a component", "component:", "- name", "include:", "local: inputs.yml"} {
		if !strings.Contains(back, want) {
			t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", want, hcl, back)
		}
	}
}

// A pipeline without a spec stays one document, so no separator is invented.
func TestNoSpecMeansNoSeparator(t *testing.T) {
	hcl, back := roundtripYAML(t, "job1:\n  script:\n    - make\n")

	if strings.Contains(back, "---") {
		t.Errorf("roundtrip invented a document separator\nHCL:\n%s\nYAML:\n%s", hcl, back)
	}

	if strings.Contains(back, "spec:") {
		t.Errorf("roundtrip invented a spec\nHCL:\n%s\nYAML:\n%s", hcl, back)
	}
}

// The same key declared in two documents is a mistake worth reporting, rather
// than one silently winning.
func TestDuplicateKeyAcrossDocumentsIsRejected(t *testing.T) {
	err := unparseYAMLString(t, "job1:\n  script:\n    - make\n---\njob1:\n  script:\n    - test\n")

	if err == nil {
		t.Fatal("Unparse() accepted a key declared in two documents")
	}

	if !strings.Contains(err.Error(), "more than one document") {
		t.Errorf("Unparse() error = %v, want it to name the duplicate", err)
	}
}
