// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A workflow and an action reach the same path through different spellings: an
// action is written to "<filename>/action.yml" and a workflow to
// "<filename>.yml", so a workflow called "build/action" lands where the action
// called "build" does. Each definition kind checked its own filenames against
// its own kind, and the keys they compared were the filenames rather than the
// paths those become, so neither saw the other. The second write landed on the
// first and the run reported success.
const collidingDefinitions = `workflow "ci" {
  filename = "build/action"

  on "push" {}

  jobs = [job.b]
}

job "b" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.a]
}

step "a" {
  run = "make"
}

action "build" {
  filename    = "build"
  name        = "Build"
  description = "builds"

  runs {
    using = "composite"
    steps = [step.a]
  }
}
`

func TestWorkflowAndActionWritingOnePathIsReported(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.hcl")

	if err := os.WriteFile(in, []byte(collidingDefinitions), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: out, YML: true})
	if !errors.Is(err, errDuplicateFilename) {
		t.Fatalf("expected errDuplicateFilename, got %v", err)
	}

	if !strings.Contains(err.Error(), "action.yml") {
		t.Errorf("the error does not name the file at fault: %v", err)
	}
}

// The same two definitions do not collide when the workflow takes the other
// extension, and both files have to be written.
func TestWorkflowAndActionOnDifferentPathsBothSurvive(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.hcl")

	if err := os.WriteFile(in, []byte(collidingDefinitions), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	if err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	for name, want := range map[string]string{
		"build/action.yaml": "jobs:",
		"build/action.yml":  "description: builds",
	} {
		got, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if !strings.Contains(string(got), want) {
			t.Errorf("%s holds the wrong definition:\n%s", name, got)
		}
	}
}
