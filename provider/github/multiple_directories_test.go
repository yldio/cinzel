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

// writeActionAt puts a composite action at <root>/<dir>/action.yml, with a
// step name of its own so two of them can be read back together.
func writeActionAt(t *testing.T, root string, dir string, name string) {
	t.Helper()

	full := filepath.Join(root, filepath.FromSlash(dir))

	if err := os.MkdirAll(full, 0o750); err != nil {
		t.Fatal(err)
	}

	action := "name: " + name + "\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: " + name + " step\n      run: echo hi\n      shell: bash\n"

	if err := os.WriteFile(filepath.Join(full, "action.yml"), []byte(action), 0o600); err != nil {
		t.Fatal(err)
	}
}

// A repository keeping composite actions in folders of its own needed one
// unparse per folder. One run reads them all, which is also what lets the
// HCL reference across them.
func TestUnparseReadsSeveralDirectories(t *testing.T) {
	tmp := t.TempDir()

	writeActionAt(t, tmp, "actions-a/setup", "Setup")
	writeActionAt(t, tmp, "actions-b/deploy", "Deploy")

	hcl := filepath.Join(tmp, "hcl")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions-a", "actions-b"},
		Recursive:       true,
		OutputDirectory: hcl,
	})
	if err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	for _, want := range []string{"setup.hcl", "deploy.hcl"} {
		if _, err := os.Stat(filepath.Join(hcl, want)); err != nil {
			t.Errorf("want %s from a single run: %v", want, err)
		}
	}
}

// With --recursive the outer directory already reaches the inner one, so the
// files underneath would be read twice: converted twice, and the second copy
// refused for sharing a name with the first. The command said something
// contradictory and is told so.
func TestOverlappingDirectoriesAreRefusedWhenRecursive(t *testing.T) {
	tmp := t.TempDir()

	writeActionAt(t, tmp, "actions-a/setup", "Setup")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions-a", "actions-a/setup"},
		Recursive:       true,
		OutputDirectory: filepath.Join(tmp, "hcl"),
	})

	if !errors.Is(err, errOverlappingDirectories) {
		t.Fatalf("Unparse() = %v, want errOverlappingDirectories", err)
	}

	for _, want := range []string{"actions-a", "actions-a/setup"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not name %q: %v", want, err)
		}
	}
}

// Without --recursive the two do not meet: "actions-a" is read shallowly, so
// the files under "actions-a/setup" are reached only by naming it.
func TestNestedDirectoriesAreFineWithoutRecursive(t *testing.T) {
	tmp := t.TempDir()

	writeActionAt(t, tmp, "actions-a/setup", "Setup")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions-a", "actions-a/setup"},
		OutputDirectory: filepath.Join(tmp, "hcl"),
	})
	if err != nil {
		t.Fatalf("Unparse() = %v, want the two read as separate sets", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "hcl", "setup.hcl")); err != nil {
		t.Errorf("want setup.hcl: %v", err)
	}
}

// Parse merges every HCL file it reads into one body, so a second root would
// be the same body spelled differently. It says so rather than quietly
// reading one of them.
func TestParseTakesOneDirectory(t *testing.T) {
	tmp := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmp, "a"), 0o750); err != nil {
		t.Fatal(err)
	}

	err := New().Parse(provider.ProviderOps{
		Directories:     []string{"a", "b"},
		OutputDirectory: filepath.Join(tmp, "out"),
	})

	if !errors.Is(err, errOneInputPathOnly) {
		t.Fatalf("Parse() = %v, want errOneInputPathOnly", err)
	}
}

// The same directory named twice is one directory, not an overlap.
func TestADirectoryNamedTwiceIsReadOnce(t *testing.T) {
	tmp := t.TempDir()

	writeActionAt(t, tmp, "actions-a/setup", "Setup")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions-a", "actions-a"},
		Recursive:       true,
		OutputDirectory: filepath.Join(tmp, "hcl"),
	})
	if err != nil {
		t.Fatalf("Unparse() = %v, want the repeat ignored", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmp, "hcl"))
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Errorf("a directory named twice produced %d files, want 1", len(entries))
	}
}
