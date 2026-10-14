// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// GitHub reads .github/workflows and nowhere else, so a run naming no
// directory means that one. Having to type the only place workflows can live
// is the flag earning nothing.
func TestUnparseDefaultsToTheWorkflowsDirectory(t *testing.T) {
	tmp := t.TempDir()

	workflows := filepath.Join(tmp, ".github", "workflows")

	if err := os.MkdirAll(workflows, 0o750); err != nil {
		t.Fatal(err)
	}

	yml := "name: CI\non:\n  push: {}\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"

	if err := os.WriteFile(filepath.Join(workflows, "ci.yaml"), []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}

	hcl := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{Root: tmp, OutputDirectory: hcl}); err != nil {
		t.Fatalf("Unparse() with no directory = %v", err)
	}

	if _, err := os.Stat(filepath.Join(hcl, "ci.hcl")); err != nil {
		t.Errorf("want the workflow converted without naming its directory: %v", err)
	}
}

// Naming it still works. The flag is redundant, not wrong, and a command
// somebody already wrote down should keep doing what it did.
func TestNamingTheWorkflowsDirectoryStillWorks(t *testing.T) {
	tmp := t.TempDir()

	workflows := filepath.Join(tmp, ".github", "workflows")

	if err := os.MkdirAll(workflows, 0o750); err != nil {
		t.Fatal(err)
	}

	yml := "name: CI\non:\n  push: {}\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"

	if err := os.WriteFile(filepath.Join(workflows, "ci.yaml"), []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}

	hcl := filepath.Join(tmp, "hcl")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directory:       ".github/workflows",
		OutputDirectory: hcl,
	})
	if err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	if _, err := os.Stat(filepath.Join(hcl, "ci.hcl")); err != nil {
		t.Errorf("want the workflow converted: %v", err)
	}
}

// The default is unparse's. Parse reads HCL, which lives wherever the author
// keeps it, so there is nothing to assume.
func TestParseDoesNotDefaultToTheWorkflowsDirectory(t *testing.T) {
	tmp := t.TempDir()

	if err := New().Parse(provider.ProviderOps{Root: tmp, OutputDirectory: filepath.Join(tmp, "out")}); err == nil {
		t.Error("Parse() with no input returned nil, want it to ask for one")
	}
}
