// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The step-only case is the last link in the detection chain and used to be
// the only one with no test of its own: whatever reached it was declared a set
// of steps and handed to the decoder. A ".github" holds a dependabot config
// and issue templates next to the workflows, and each of those came back as
// "not a valid type", which aborted the whole directory before the workflows
// beside them were reached.
const dependabotConfig = `version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
`

const issueTemplate = `name: Bug report
description: File a bug
labels: ["bug"]
body:
  - type: textarea
    id: what
    attributes:
      label: What happened?
`

func TestDocumentThatIsNoneOfTheThreeTypesIsSkipped(t *testing.T) {
	err, out := unparseDir(t, map[string]string{
		"ci.yml":         realWorkflow,
		"dependabot.yml": dependabotConfig,
		"bug.yml":        issueTemplate,
	})
	if err != nil {
		t.Fatalf("a file this tool has no definition for should be skipped, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(out, "real.hcl")); statErr == nil {
		t.Fatal("test wrote the wrong name")
	}

	if _, statErr := os.Stat(filepath.Join(out, "ci.hcl")); statErr != nil {
		t.Errorf("the workflow beside them was not converted: %v", statErr)
	}

	for _, name := range []string{"dependabot.hcl", "bug.hcl"} {
		if _, statErr := os.Stat(filepath.Join(out, name)); !os.IsNotExist(statErr) {
			t.Errorf("%s should not have produced HCL", name)
		}
	}
}

// Named on its own, such a file converts to nothing, which the end-of-run
// guard reports. "not a valid type" said neither what was wrong nor which
// field it meant.
func TestDocumentThatIsNoneOfTheThreeTypesIsReportedWhenNamed(t *testing.T) {
	err, _ := unparseDir(t, map[string]string{"dependabot.yml": dependabotConfig})

	if !errors.Is(err, errNoDefinitions) {
		t.Fatalf("expected errNoDefinitions, got %v", err)
	}
}

// A real step-only file has to keep converting, and a broken step inside one
// has to keep being an error rather than a silent skip.
func TestStepOnlyDocumentIsStillRecognised(t *testing.T) {
	err, out := unparseDir(t, map[string]string{"steps.yml": "checkout:\n  uses: actions/checkout@v4\nbuild:\n  run: make\n"})
	if err != nil {
		t.Fatalf("a step-only file should still convert: %v", err)
	}

	got, readErr := os.ReadFile(filepath.Join(out, "steps.hcl"))
	if readErr != nil {
		t.Fatal(readErr)
	}

	if !strings.Contains(string(got), `step "checkout"`) {
		t.Errorf("step block missing from output:\n%s", got)
	}
}

func TestBrokenStepInsideAStepOnlyDocumentIsStillAnError(t *testing.T) {
	err, _ := unparseDir(t, map[string]string{"steps.yml": "build:\n  run: make\n  shell: [not, a, string]\n"})
	if err == nil {
		t.Fatal("a step that cannot be decoded should still be an error")
	}
}
