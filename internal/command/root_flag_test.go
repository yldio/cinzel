// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const rootWorkflow = `step "echo" {
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

// --root moves a whole run: what it reads, and what it writes. One setting
// rather than one per path is what makes it usable for a scratch checkout or
// a test that wants the entire output shape in one place.
func TestRootRelocatesTheRun(t *testing.T) {
	t.Run("parse writes under the root", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			if err := os.MkdirAll(filepath.Join("src", "hcl"), 0o750); err != nil {
				t.Fatal(err)
			}

			writeFile(t, filepath.Join("src", "hcl", "ci.hcl"), []byte(rootWorkflow))

			if err := runCommandLine(t, []string{
				"cinzel", "github", "parse", "--root", "src", "--directory", "hcl",
			}); err != nil {
				t.Fatalf("parse --root = %v", err)
			}

			if _, err := os.Stat(filepath.Join("src", ".github", "workflows", "ci.yaml")); err != nil {
				t.Errorf("want the workflow under the root: %v", err)
			}

			// Nothing outside the root: that is the whole promise.
			if _, err := os.Stat(filepath.Join(".github", "workflows", "ci.yaml")); !os.IsNotExist(err) {
				t.Error("a run given a root wrote outside it")
			}
		})
	})

	t.Run("unparse reads and writes under the root", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			if err := os.MkdirAll(filepath.Join("src", ".github", "workflows"), 0o750); err != nil {
				t.Fatal(err)
			}

			yml := "name: CI\non:\n  push: {}\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"
			writeFile(t, filepath.Join("src", ".github", "workflows", "ci.yaml"), []byte(yml))

			if err := runCommandLine(t, []string{
				"cinzel", "github", "unparse", "--root", "src", "--directory", ".github/workflows",
			}); err != nil {
				t.Fatalf("unparse --root = %v", err)
			}

			if _, err := os.Stat(filepath.Join("src", "cinzel", "ci.hcl")); err != nil {
				t.Errorf("want the HCL under the root: %v", err)
			}
		})
	})
}

// The two answer one question. A precedence rule would be a thing to
// remember; the error says which to use instead.
func TestRootAndOutputDirectoryTogetherAreRefused(t *testing.T) {
	for _, command := range []string{"parse", "unparse"} {
		t.Run(command, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				err := runCommandLine(t, []string{
					"cinzel", "github", command,
					"--root", "src", "--output-directory", "out", "--directory", "in",
				})

				if !errors.Is(err, errRootAndOutputDirectory) {
					t.Fatalf("%s with both flags = %v, want errRootAndOutputDirectory", command, err)
				}

				if !strings.Contains(err.Error(), "--root") {
					t.Errorf("the error does not name the flags: %v", err)
				}
			})
		})
	}
}

// A root is a path like any other cinzel reads, so it cannot climb out of the
// working directory any more than a configured one can.
func TestRootCannotEscapeTheWorkingDirectory(t *testing.T) {
	withTempWorkingDir(t, func() {
		for _, root := range []string{"../elsewhere", "../../etc"} {
			err := runCommandLine(t, []string{
				"cinzel", "github", "parse", "--root", root, "--directory", "hcl",
			})

			if !errors.Is(err, errPathTraversal) {
				t.Errorf("--root %q = %v, want it refused", root, err)
			}
		}
	})
}
