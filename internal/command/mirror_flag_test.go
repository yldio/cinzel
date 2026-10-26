// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const mirrorAction = "name: Setup\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: Setup step\n      run: npm ci\n      shell: bash\n"

const mirrorWorkflow = "name: CI\non:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n"

// writeMirrorRepo lays out a workflow and an action in a folder of its own.
func writeMirrorRepo(t *testing.T) {
	t.Helper()

	for _, dir := range []string{filepath.Join(".github", "workflows"), filepath.Join("actions", "setup")} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
	}

	writeFile(t, filepath.Join(".github", "workflows", "ci.yaml"), []byte(mirrorWorkflow))
	writeFile(t, filepath.Join("actions", "setup", "action.yml"), []byte(mirrorAction))
}

// The command that was reported broken: it converted the actions and dropped
// every workflow in the repository.
func TestUnparseWithADirectoryStillTakesTheWorkflows(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeMirrorRepo(t)

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "-d", "actions", "-r", "--output-directory", "cinzel",
		}); err != nil {
			t.Fatalf("unparse = %v", err)
		}

		for _, want := range []string{"ci.hcl", "setup.hcl"} {
			if _, err := os.Stat(filepath.Join("cinzel", want)); err != nil {
				t.Errorf("want cinzel/%s: %v", want, err)
			}
		}
	})
}

// --mirror reaches the provider, and the tree it writes parses back with
// --recursive.
func TestMirrorFlagRoundTrip(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeMirrorRepo(t)

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "-d", "actions", "-r", "--mirror", "--output-directory", "cinzel",
		}); err != nil {
			t.Fatalf("unparse --mirror = %v", err)
		}

		// The workflow is flat even here: ".github/workflows" is the one
		// place GitHub reads, so it is not a layout to reproduce, and parse
		// writes every workflow back there wherever the HCL sits.
		for _, want := range []string{"ci.hcl", "actions/setup.hcl"} {
			if _, err := os.Stat(filepath.Join("cinzel", filepath.FromSlash(want))); err != nil {
				t.Errorf("want cinzel/%s: %v", want, err)
			}
		}

		if err := os.RemoveAll(".github"); err != nil {
			t.Fatal(err)
		}

		if err := os.RemoveAll("actions"); err != nil {
			t.Fatal(err)
		}

		if err := runCommandLine(t, []string{"cinzel", "github", "parse", "-d", "cinzel", "-r"}); err != nil {
			t.Fatalf("parse of a mirrored tree = %v", err)
		}

		for _, want := range []string{".github/workflows/ci.yaml", "actions/setup/action.yml"} {
			if _, err := os.Stat(filepath.FromSlash(want)); err != nil {
				t.Errorf("want %s back: %v", want, err)
			}
		}
	})
}

// Without --mirror the output stays flat, which is what was released.
func TestWithoutTheMirrorFlagTheOutputIsFlat(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeMirrorRepo(t)

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "-d", "actions", "-r", "--output-directory", "cinzel",
		}); err != nil {
			t.Fatalf("unparse = %v", err)
		}

		entries, err := os.ReadDir("cinzel")
		if err != nil {
			t.Fatal(err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				t.Errorf("a flat conversion made the directory cinzel/%s", entry.Name())
			}
		}
	})
}

// A mirrored tree read without --recursive used to say there was no HCL at
// all, which reads as "your files are missing" rather than "add a flag".
//
// The repository here holds an action and no workflow, so the mirrored output
// has nothing at its top level. With a workflow in it the top level is no
// longer empty — a workflow is written flat — and parse reads that one file
// and reports no error, which is the right answer for what it was given.
func TestParsingAMirroredTreeWithoutRecursiveSaysWhatToDo(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeMirrorRepo(t)

		if err := os.RemoveAll(".github"); err != nil {
			t.Fatal(err)
		}

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "-d", "actions", "-r", "--mirror", "--output-directory", "cinzel",
		}); err != nil {
			t.Fatal(err)
		}

		err := runCommandLine(t, []string{"cinzel", "github", "parse", "-d", "cinzel"})
		if err == nil {
			t.Fatal("parse of a mirrored tree without --recursive returned nil")
		}

		if !strings.Contains(err.Error(), "--recursive") {
			t.Errorf("the error does not name the flag that fixes it: %v", err)
		}
	})
}
