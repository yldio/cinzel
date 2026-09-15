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

// actionYAML is one composite action, named so the two written in a single
// run can be told apart.
func actionYAML(name string) string {
	return "name: " + name + "\ndescription: d\nruns:\n  using: composite\n  steps:\n    - id: s" + name + "\n      run: echo " + name + "\n"
}

// writeAction puts an action where GitHub reads one from: its own directory,
// under the fixed filename.
func writeAction(t *testing.T, root, dir, filename string) {
	t.Helper()

	full := filepath.Join(root, dir)
	if err := os.MkdirAll(full, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(full, filename), []byte(actionYAML(dir)), 0o600); err != nil {
		t.Fatal(err)
	}
}

// An action is identified by the directory holding it, because every action
// file is called the same thing. Naming the output after the file instead
// called it "action", so the action came back under a name nobody chose.
func TestUnparsedActionKeepsItsDirectoryName(t *testing.T) {
	for _, filename := range []string{"action.yml", "action.yaml"} {
		t.Run(filename, func(t *testing.T) {
			dir := t.TempDir()
			in := filepath.Join(dir, "in")
			out := filepath.Join(dir, "out")

			writeAction(t, in, "alpha", filename)

			if err := New().Unparse(provider.ProviderOps{Directory: in, Recursive: true, OutputDirectory: out}); err != nil {
				t.Fatal(err)
			}

			hcl, err := os.ReadFile(filepath.Join(out, "alpha.hcl"))
			if err != nil {
				t.Fatalf("want the output named after the directory, got %v", err)
			}

			for _, want := range []string{`action "alpha"`, `filename = "alpha"`} {
				if !strings.Contains(string(hcl), want) {
					t.Errorf("want %s in the output, got:\n%s", want, hcl)
				}
			}
		})
	}
}

// Two actions are two files. They share a basename, so naming the output
// after it left one file holding whichever was read last.
func TestTwoActionsUnparseToTwoFiles(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in")
	out := filepath.Join(dir, "out")

	writeAction(t, in, "alpha", "action.yml")
	writeAction(t, in, "beta", "action.yml")

	if err := New().Unparse(provider.ProviderOps{Directory: in, Recursive: true, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"alpha", "beta"} {
		if _, err := os.Stat(filepath.Join(out, name+".hcl")); err != nil {
			t.Errorf("want %s.hcl kept, got %v", name, err)
		}
	}
}

// The name has to survive both directions, or the action is written back to
// a different place than it was read from.
func TestActionRoundTripsThroughItsDirectory(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in")
	hclDir := filepath.Join(dir, "hcl")
	back := filepath.Join(dir, "back")

	writeAction(t, in, "alpha", "action.yml")

	if err := New().Unparse(provider.ProviderOps{Directory: in, Recursive: true, OutputDirectory: hclDir}); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: filepath.Join(hclDir, "alpha.hcl"), OutputDirectory: back}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(back, "alpha", "action.yml")); err != nil {
		t.Errorf("want the action back under 'alpha', got %v", err)
	}
}

// A workflow is named by its own file, and an action sitting under some
// other name has nothing else to go on. Neither may be renamed by this.
func TestOtherDocumentsKeepTheirFilename(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in")
	out := filepath.Join(dir, "out")

	if err := os.MkdirAll(in, 0o750); err != nil {
		t.Fatal(err)
	}

	workflow := "name: W\non:\n  push:\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - id: s\n        run: echo hi\n"
	if err := os.WriteFile(filepath.Join(in, "ci.yaml"), []byte(workflow), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(in, "renamed.yml"), []byte(actionYAML("renamed")), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{Directory: in, Recursive: true, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"ci", "renamed"} {
		if _, err := os.Stat(filepath.Join(out, name+".hcl")); err != nil {
			t.Errorf("want %s.hcl kept, got %v", name, err)
		}
	}
}
