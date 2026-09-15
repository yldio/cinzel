// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/internal/fsutil"
	"github.com/yldio/cinzel/provider"
)

// actionOnlyHCL is one composite action under the given name.
func actionOnlyHCL(id string) string {
	return `step "hi" {
  run = "echo hi"
}
action "` + id + `" {
  filename = "` + id + `"
  name = "A"
  description = "d"
  runs {
    using = "composite"
    steps = [step.hi]
  }
}
`
}

// parseActionHCL writes the HCL and parses it into output.
func parseActionHCL(t *testing.T, dir, output, id string) {
	t.Helper()

	path := filepath.Join(dir, "in.hcl")
	if err := os.WriteFile(path, []byte(actionOnlyHCL(id)), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: output}); err != nil {
		t.Fatal(err)
	}
}

// An action is written with the same marker a workflow gets. Without it
// nothing can tell an action cinzel wrote from one written by hand.
func TestActionIsWrittenWithTheGeneratedMarker(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out")

	parseActionHCL(t, dir, output, "alpha")

	written := filepath.Join(output, "alpha", "action.yml")

	owned, err := fsutil.HasGeneratedMarker(written, providerName)
	if err != nil {
		t.Fatal(err)
	}

	if !owned {
		got, readErr := os.ReadFile(written)
		if readErr != nil {
			t.Fatal(readErr)
		}

		t.Fatalf("want the action marked, got:\n%s", got)
	}
}

// Renaming an action left the old one behind with a live action.yml in it,
// because the prune could not tell it was cinzel's to remove.
func TestRenamedActionIsPruned(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out")

	parseActionHCL(t, dir, output, "before")
	parseActionHCL(t, dir, output, "after")

	if _, err := os.Stat(filepath.Join(output, "before", "action.yml")); !os.IsNotExist(err) {
		t.Errorf("want the renamed-away action removed, stat err=%v", err)
	}

	if _, err := os.Stat(filepath.Join(output, "after", "action.yml")); err != nil {
		t.Errorf("want the current action kept, got %v", err)
	}
}

// The prune only removes what carries the marker, so an action written by
// hand in the same directory is not cinzel's to delete.
func TestHandWrittenActionSurvivesThePrune(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out")

	manual := filepath.Join(output, "manual")
	if err := os.MkdirAll(manual, 0o750); err != nil {
		t.Fatal(err)
	}

	body := "name: M\ndescription: d\nruns:\n  using: composite\n  steps: []\n"
	if err := os.WriteFile(filepath.Join(manual, "action.yml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	parseActionHCL(t, dir, output, "generated")

	if _, err := os.Stat(filepath.Join(manual, "action.yml")); err != nil {
		t.Errorf("want the hand-written action kept, got %v", err)
	}
}

// A marked action has to read back, or marking it breaks the other
// direction: the comments are dropped and the action keeps its name.
func TestMarkedActionUnparsesBack(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out")
	back := filepath.Join(dir, "back")

	parseActionHCL(t, dir, output, "alpha")

	if err := New().Unparse(provider.ProviderOps{Directory: output, Recursive: true, OutputDirectory: back}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(back, "alpha.hcl"))
	if err != nil {
		t.Fatalf("want the action back under its own name, got %v", err)
	}

	if strings.Contains(string(got), "generated-by") {
		t.Errorf("want the marker dropped on the way back, got:\n%s", got)
	}
}
