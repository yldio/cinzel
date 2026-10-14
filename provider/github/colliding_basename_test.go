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

// Two files converting to one workflow name is refused, and the run writes
// nothing.
//
// This asserted the opposite until it was found that what it blessed does not
// read back. The bug it was written for was real: the second file overwrote
// the first and said nothing, and UniqueOutputName fixed that by writing
// "ci.hcl" and "ci_2.hcl". What went unchecked is the "filename" attribute
// inside them, which is the workflow's name on GitHub rather than a block
// label and so is not suffixed. Both files came out saying filename = "ci",
// and parse refused to read its own output:
//
//	two definitions write to the same file: 'ci' and 'ci' both write to 'ci'
//
// So the run produced two files, lost nothing, and left the user holding HCL
// cinzel would not take back — one command later, with neither YAML file
// named. Suffixing the attribute too would roundtrip, at the cost of renaming
// someone's workflow on GitHub: a badge, a branch protection rule and a
// workflow_run trigger all point at that name.
func TestCollidingWorkflowNamesAreRefused(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in")
	out := filepath.Join(tmp, "out")

	for _, dir := range []string{"a", "b"} {
		path := filepath.Join(in, dir)

		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}

		yml := "name: from-" + dir + "\non:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo " + dir + "\n"

		if err := os.WriteFile(filepath.Join(path, "ci.yaml"), []byte(yml), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	err := New().Unparse(provider.ProviderOps{Directory: in, Recursive: true, OutputDirectory: out})

	if !errors.Is(err, errUnparseNameClash) {
		t.Fatalf("Unparse() = %v, want errUnparseNameClash", err)
	}

	// Both inputs named, or the user has to find them by hand.
	for _, want := range []string{"ci.yaml", "'ci'"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not mention %q: %v", want, err)
		}
	}

	// A refused run leaves the output directory as it was, so a second
	// attempt starts from where the first did.
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		entries, readErr := os.ReadDir(out)
		if readErr != nil {
			t.Fatal(readErr)
		}

		if len(entries) != 0 {
			t.Errorf("a refused run wrote %d files", len(entries))
		}
	}
}

// Two files whose names do not collide still both convert, which is the half
// of the original bug that stands: neither overwrites the other.
func TestDistinctBasenamesBothConvert(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in")
	out := filepath.Join(tmp, "out")

	for _, name := range []string{"ci", "cd"} {
		path := filepath.Join(in, name)

		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}

		yml := "name: from-" + name + "\non:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo " + name + "\n"

		if err := os.WriteFile(filepath.Join(path, name+".yaml"), []byte(yml), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if err := New().Unparse(provider.ProviderOps{Directory: in, Recursive: true, OutputDirectory: out}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatalf("want 2 files, got %d", len(entries))
	}

	var joined string

	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(out, e.Name()))
		if err != nil {
			t.Fatal(err)
		}

		joined += string(b)
	}

	for _, want := range []string{"from-ci", "from-cd"} {
		if !strings.Contains(joined, want) {
			t.Errorf("want %q in the output, got:\n%s", want, joined)
		}
	}
}

// A run with nothing to collide keeps the name it took from the input.
func TestASingleFileKeepsItsName(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "ci.yaml")
	out := filepath.Join(tmp, "out")

	yml := "name: only\non:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"

	if err := os.WriteFile(in, []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "ci.hcl")); err != nil {
		t.Errorf("want ci.hcl, got %v", err)
	}
}
