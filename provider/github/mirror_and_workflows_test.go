// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// repoWithWorkflowAndAction lays out the shape this is all about: a workflow
// where GitHub reads one, and a composite action somewhere else.
func repoWithWorkflowAndAction(t *testing.T) string {
	t.Helper()

	tmp := t.TempDir()

	workflows := filepath.Join(tmp, ".github", "workflows")
	action := filepath.Join(tmp, "actions", "setup")

	for _, dir := range []string{workflows, action} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
	}

	yml := "name: CI\non:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: make\n"
	if err := os.WriteFile(filepath.Join(workflows, "ci.yaml"), []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}

	act := "name: Setup\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: Setup step\n      run: npm ci\n      shell: bash\n"
	if err := os.WriteFile(filepath.Join(action, "action.yml"), []byte(act), 0o600); err != nil {
		t.Fatal(err)
	}

	return tmp
}

// "unparse -d actions -r" converted the actions and dropped every workflow in
// the repository. The workflows directory was added only to a run naming no
// directory at all, so the one directory nobody should have to type became
// the one directory naming anything else threw away.
func TestUnparseReadsTheWorkflowsDirectoryAlongsideOthers(t *testing.T) {
	tmp := repoWithWorkflowAndAction(t)
	hcl := filepath.Join(tmp, "cinzel")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions"},
		Recursive:       true,
		OutputDirectory: hcl,
	})
	if err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	for _, want := range []string{"ci.hcl", "setup.hcl"} {
		if _, err := os.Stat(filepath.Join(hcl, want)); err != nil {
			t.Errorf("want %s: %v", want, err)
		}
	}
}

// A repository with composite actions and no workflows at all is an ordinary
// thing to have. The directory is cinzel's assumption, not the caller's
// instruction, so its absence is not their mistake.
func TestUnparseWithoutAWorkflowsDirectory(t *testing.T) {
	tmp := t.TempDir()

	action := filepath.Join(tmp, "actions", "setup")
	if err := os.MkdirAll(action, 0o750); err != nil {
		t.Fatal(err)
	}

	act := "name: Setup\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: Setup step\n      run: npm ci\n      shell: bash\n"
	if err := os.WriteFile(filepath.Join(action, "action.yml"), []byte(act), 0o600); err != nil {
		t.Fatal(err)
	}

	hcl := filepath.Join(tmp, "cinzel")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions"},
		Recursive:       true,
		OutputDirectory: hcl,
	})
	if err != nil {
		t.Fatalf("Unparse() = %v, want the missing directory passed over", err)
	}

	if _, err := os.Stat(filepath.Join(hcl, "setup.hcl")); err != nil {
		t.Errorf("want setup.hcl: %v", err)
	}
}

// --mirror lays the HCL out the way the repository is laid out. Where a file
// goes on the way back is its own "path" attribute either way, so this is for
// reading the tree rather than for the conversion.
func TestMirrorLaysTheHCLOutLikeTheRepository(t *testing.T) {
	tmp := repoWithWorkflowAndAction(t)
	hcl := filepath.Join(tmp, "cinzel")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions"},
		Recursive:       true,
		Mirror:          true,
		OutputDirectory: hcl,
	})
	if err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	for _, want := range []string{".github/workflows/ci.hcl", "actions/setup.hcl"} {
		if _, err := os.Stat(filepath.Join(hcl, filepath.FromSlash(want))); err != nil {
			t.Errorf("want %s: %v", want, err)
		}
	}

	// And back out again, which a mirrored tree needs --recursive for.
	//
	// Read against a root rather than an output directory, so each file
	// returns where it came from: --output-directory is the workflow
	// destination, which would put ci.yaml at its top level instead.
	back := filepath.Join(tmp, "back")

	if err := os.MkdirAll(back, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{Root: back, Directory: hcl, Recursive: true}); err != nil {
		t.Fatalf("Parse() = %v", err)
	}

	for _, want := range []string{".github/workflows/ci.yaml", "actions/setup/action.yml"} {
		if _, err := os.Stat(filepath.Join(back, filepath.FromSlash(want))); err != nil {
			t.Errorf("want %s back: %v", want, err)
		}
	}
}

// Without --mirror everything stays in one folder, which is the default and
// what was released.
func TestWithoutMirrorTheHCLIsFlat(t *testing.T) {
	tmp := repoWithWorkflowAndAction(t)
	hcl := filepath.Join(tmp, "cinzel")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directories:     []string{"actions"},
		Recursive:       true,
		OutputDirectory: hcl,
	})
	if err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	entries, err := os.ReadDir(hcl)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Errorf("a flat conversion made the directory %s", entry.Name())
		}
	}
}
