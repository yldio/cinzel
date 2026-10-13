// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const diffWorkflow = `step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "ci"

  on "push" {}

  jobs = [job.build]
}
`

// writeDiffTree lays out an hcl directory and parses it into out, so the two
// start in step.
func writeDiffTree(t *testing.T) {
	t.Helper()

	if err := os.MkdirAll("hcl", 0o750); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join("hcl", "ci.hcl"), []byte(diffWorkflow))

	if err := runCommandLine(t, []string{
		"cinzel", "github", "parse", "--directory", "hcl", "--output-directory", "out",
	}); err != nil {
		t.Fatal(err)
	}
}

// The three outcomes a workflow branches on. Drift and a cinzel that could not
// run are both a red step, and telling them apart is the whole reason the
// second code exists.
func TestDiffExitCodes(t *testing.T) {
	t.Run("a tree in step returns nil", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			writeDiffTree(t)

			if err := runCommandLine(t, []string{
				"cinzel", "github", "diff", "--directory", "hcl", "--output-directory", "out",
			}); err != nil {
				t.Errorf("diff of an unchanged tree = %v, want nil", err)
			}
		})
	})

	t.Run("drift returns ErrDiffFound", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			writeDiffTree(t)

			writeFile(t, filepath.Join("out", "ci.yaml"), []byte("name: edited by hand\n"))

			err := runCommandLine(t, []string{
				"cinzel", "github", "diff", "--directory", "hcl", "--output-directory", "out",
			})

			if !errors.Is(err, ErrDiffFound) {
				t.Errorf("diff of a drifted tree = %v, want ErrDiffFound", err)
			}

			if errors.Is(err, ErrCannotRun) {
				t.Error("drift was reported as a run that failed, which exits 2")
			}
		})
	})

	t.Run("input it cannot read returns ErrCannotRun", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			err := runCommandLine(t, []string{
				"cinzel", "github", "diff", "--directory", "absent", "--output-directory", "out",
			})

			if !errors.Is(err, ErrCannotRun) {
				t.Errorf("diff of a missing directory = %v, want ErrCannotRun", err)
			}

			if errors.Is(err, ErrDiffFound) {
				t.Error("a run that failed was reported as drift, which exits 1")
			}
		})
	})

	// Every other command keeps exiting 1, so none of them may carry the
	// mark main reads as 2.
	t.Run("another command's failure is not ErrCannotRun", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			err := runCommandLine(t, []string{
				"cinzel", "github", "parse", "--directory", "absent",
			})

			if err == nil {
				t.Fatal("parse of a missing directory returned nil")
			}

			if errors.Is(err, ErrCannotRun) {
				t.Error("parse now exits 2, which changes a code it has always returned")
			}
		})
	})
}
