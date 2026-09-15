// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// unparseStepsYAML runs Unparse over a step-only YAML file and returns the HCL.
func unparseStepsYAML(t *testing.T, yaml string) string {
	t.Helper()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "steps.yaml")

	if err := os.WriteFile(in, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatalf("unparse: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(out, "steps.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	return string(got)
}

// Steps in a step-only file rarely carry the same keys: one runs a command,
// the next uses an action. Reading the document as a map of a single element
// type made cty unify them, and two steps shaped differently were rejected
// with "not a valid type".
func TestStepOnlyStepsMayDifferInKeys(t *testing.T) {
	got := unparseStepsYAML(t, `checkout:
  uses: actions/checkout@v4
build:
  run: make
  working-directory: ./src
`)

	for _, want := range []string{`step "checkout"`, `step "build"`, `action  = "actions/checkout"`, `run = "make"`, `working_directory = "./src"`} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in output, got:\n%s", want, got)
		}
	}
}

// A step carrying keys the others lack, in a longer file, is the same bug at
// a size where the unification failure is easier to miss.
func TestStepOnlyMixedShapesSurvive(t *testing.T) {
	got := unparseStepsYAML(t, `one:
  run: echo one
two:
  run: echo two
  shell: bash
three:
  run: echo three
  continue-on-error: true
  timeout-minutes: 5
`)

	for _, want := range []string{"shell = \"bash\"", "continue_on_error = true", "timeout_minutes = 5"} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in output, got:\n%s", want, got)
		}
	}
}

// Steps sharing one shape kept working before and have to keep working now.
func TestStepOnlyUniformStepsStillWork(t *testing.T) {
	got := unparseStepsYAML(t, `one:
  run: echo one
two:
  run: echo two
`)

	for _, want := range []string{`step "one"`, `step "two"`} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in output, got:\n%s", want, got)
		}
	}
}
