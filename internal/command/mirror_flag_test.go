// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
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

// A composite action is always at "<dir>/<name>/action.yml", one directory
// below the folder holding it, so a folder of actions holds no YAML of its
// own. Reading only its top level converted none of them: the run wrote the
// workflows it adds on its own and reported success, so the directory the
// caller actually named looked as though it had been read.
//
// Naming it is asking for what is in it, so --recursive is not needed too.
func TestADirectoryOfActionsIsReadWithoutRecursive(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeMirrorRepo(t)

		actions := filepath.Join(".github", "actions", "setup")

		if err := os.MkdirAll(actions, 0o750); err != nil {
			t.Fatal(err)
		}

		writeFile(t, filepath.Join(actions, "action.yml"), []byte(mirrorAction))

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "--mirror",
			"-d", filepath.Join(".github", "actions"),
			"--output-directory", "cinzel",
		}); err != nil {
			t.Fatalf("unparse = %v", err)
		}

		for _, want := range []string{".github/actions/setup.hcl", "ci.hcl"} {
			if _, err := os.Stat(filepath.Join("cinzel", filepath.FromSlash(want))); err != nil {
				t.Errorf("want cinzel/%s: %v", want, err)
			}
		}
	})
}

// Only one level down, and only an action file. A directory of directories
// keeps everything else there too, and a caller who means all of it has
// --recursive to say so.
func TestOnlyAnActionIsTakenFromTheLevelBelow(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeMirrorRepo(t)

		deep := filepath.Join(".github", "actions", "deep", "nested")

		if err := os.MkdirAll(deep, 0o750); err != nil {
			t.Fatal(err)
		}

		writeFile(t, filepath.Join(deep, "action.yml"), []byte(mirrorAction))

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "--mirror",
			"-d", filepath.Join(".github", "actions"),
			"--output-directory", "cinzel",
		}); err != nil {
			t.Fatalf("unparse = %v", err)
		}

		// Two levels down is past where an action of this directory lives.
		if _, err := os.Stat(filepath.Join("cinzel", ".github", "actions", "deep")); err == nil {
			t.Error("a file two directories down was read without --recursive")
		}
	})
}

// With --recursive the same run finds the action, and the workflow beside it
// still lands flat.
func TestTheSameDirectoryWithRecursiveConvertsTheAction(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeMirrorRepo(t)

		actions := filepath.Join(".github", "actions", "setup")

		if err := os.MkdirAll(actions, 0o750); err != nil {
			t.Fatal(err)
		}

		writeFile(t, filepath.Join(actions, "action.yml"), []byte(mirrorAction))

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "--mirror", "-r",
			"-d", filepath.Join(".github", "actions"),
			"--output-directory", "cinzel",
		}); err != nil {
			t.Fatalf("unparse = %v", err)
		}

		for _, want := range []string{"ci.hcl", ".github/actions/setup.hcl"} {
			if _, err := os.Stat(filepath.Join("cinzel", filepath.FromSlash(want))); err != nil {
				t.Errorf("want cinzel/%s: %v", want, err)
			}
		}
	})
}

// --ignore-unknown reaches the provider, and the file is refused without it.
func TestIgnoreUnknownFlagReachesTheProvider(t *testing.T) {
	const workflow = `name: ci
on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    bogus: x
    steps:
      - run: make
`

	withTempWorkingDir(t, func() {
		if err := os.MkdirAll(filepath.Join(".github", "workflows"), 0o750); err != nil {
			t.Fatal(err)
		}

		writeFile(t, filepath.Join(".github", "workflows", "ci.yaml"), []byte(workflow))

		err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "--output-directory", "cinzel",
		})

		if err == nil {
			t.Fatal("the unknown key was not refused by default")
		}

		if err := runCommandLine(t, []string{
			"cinzel", "github", "unparse", "--ignore-unknown", "--output-directory", "cinzel",
		}); err != nil {
			t.Fatalf("unparse --ignore-unknown = %v", err)
		}

		if _, err := os.Stat(filepath.Join("cinzel", "ci.hcl")); err != nil {
			t.Errorf("want cinzel/ci.hcl: %v", err)
		}
	})
}

// It reads from the configuration file too, and the flag still wins.
func TestIgnoreUnknownInTheConfigFile(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  unparse:\n    ignore-unknown: true\n"))

		app, _, p := newConfigTestApp(t)

		if err := app.Execute([]string{"cinzel", "github", "unparse"}, []provider.Provider{p}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if !p.unparseOpts.IgnoreUnknown {
			t.Fatal("ignore-unknown = false, want the file's true")
		}

		app, _, p = newConfigTestApp(t)

		if err := app.Execute([]string{"cinzel", "github", "unparse", "--ignore-unknown=false"}, []provider.Provider{p}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.unparseOpts.IgnoreUnknown {
			t.Fatal("ignore-unknown = true, want the flag to have won")
		}
	})
}
