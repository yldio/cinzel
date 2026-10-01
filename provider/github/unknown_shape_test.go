// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A mapping of mappings is the shape of a step-only file, and the step decoder
// reads the keys it knows and ignores the rest. So any document of that shape
// converted: this one came out as a step named "something", with "do" dropped
// and nothing said, at exit 0.
const notAStepDocument = `something:
  do: another
`

func TestAMappingOfMappingsIsNotAStepFile(t *testing.T) {
	err, out := unparseDir(t, map[string]string{
		"ci.yml":   realWorkflow,
		"file.yml": notAStepDocument,
	})
	if err != nil {
		t.Fatalf("a file this tool has no definition for should be skipped, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(out, "file.hcl")); !os.IsNotExist(statErr) {
		got, _ := os.ReadFile(filepath.Join(out, "file.hcl"))
		t.Errorf("a document that is not a set of steps was converted anyway:\n%s", got)
	}

	if _, statErr := os.Stat(filepath.Join(out, "ci.hcl")); statErr != nil {
		t.Errorf("the workflow beside it was not converted: %v", statErr)
	}
}

// Named on its own it converts to nothing, which the end-of-run guard reports
// rather than leaving the run to exit 0 having written no file.
func TestAMappingOfMappingsNamedAloneIsReported(t *testing.T) {
	err, _ := unparseDir(t, map[string]string{"file.yml": notAStepDocument})

	if !errors.Is(err, errNoDefinitions) {
		t.Fatalf("expected errNoDefinitions, got %v", err)
	}
}

// A file skipped in the middle of a directory run left no trace, so a run over
// a ".github" holding one workflow and four other YAML files reported the same
// success as one that converted all five.
func TestASkippedFileIsWarnedAbout(t *testing.T) {
	var out string

	warnings := captureGitHubStderr(t, func() {
		err, dir := unparseDir(t, map[string]string{
			"ci.yml":   realWorkflow,
			"file.yml": notAStepDocument,
		})
		if err != nil {
			t.Errorf("unparse: %v", err)
		}

		out = dir
	})

	if !strings.Contains(warnings, "file.yml") {
		t.Errorf("the skipped file was not named in a warning, got: %q", warnings)
	}

	if strings.Contains(warnings, "ci.yml") {
		t.Errorf("the converted file should not be warned about, got: %q", warnings)
	}

	if _, err := os.Stat(filepath.Join(out, "ci.hcl")); err != nil {
		t.Errorf("the workflow was not converted: %v", err)
	}
}

// A step is a mapping whose every key is one a step declares. The keys a step
// does not declare are exactly what the decoder would drop, so a document
// carrying one is not a set of steps.
//
// Carrying no key that runs anything is not the test. Parse writes a step with
// only a name, and refusing that here refused cinzel's own output.
func TestOnlyDocumentsShapedLikeStepsConvert(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		converts bool
	}{
		{"a step that runs a command", "build:\n  run: make\n", true},
		{"a step that uses an action", "checkout:\n  uses: actions/checkout@v4\n", true},
		{"a step with every key it may carry", "build:\n  id: build\n  name: Build\n  if: always()\n  run: make\n  shell: bash\n  working-directory: ./src\n  env:\n    FOO: bar\n  continue-on-error: true\n  timeout-minutes: 5\n", true},
		{"a mapping carrying a key no step declares", "something:\n  do: another\n", false},
		{"a step carrying only a name, which is what parse writes", "something:\n  name: a name\n", true},
		{"a step key beside one no step declares", "something:\n  run: make\n  do: another\n", false},
		{"an empty mapping", "something: {}\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, out := unparseDir(t, map[string]string{"doc.yml": tt.yaml})

			if tt.converts {
				if err != nil {
					t.Fatalf("expected a step-only file to convert, got %v", err)
				}

				if _, statErr := os.Stat(filepath.Join(out, "doc.hcl")); statErr != nil {
					t.Fatalf("no output written: %v", statErr)
				}

				return
			}

			if !errors.Is(err, errNoDefinitions) {
				t.Fatalf("expected the document to be skipped, got %v", err)
			}
		})
	}
}

// captureGitHubStderr returns whatever fn wrote to os.Stderr. The warning goes
// there rather than into an error, so there is nothing else to assert on.
func captureGitHubStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	original := os.Stderr
	os.Stderr = w

	defer func() { os.Stderr = original }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return string(out)
}
