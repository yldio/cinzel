// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// --file names a file. It used to take a directory as well, because the path
// it returned was stat'd again further down and the directory branch picked it
// up — undeclared, untested, and the opposite of what the flag says. A reader
// reaching for it had --directory sitting beside it for exactly that.
func TestFileFlagRefusesADirectory(t *testing.T) {
	tmp := t.TempDir()

	hcl := filepath.Join(tmp, "ci.hcl")
	if err := os.WriteFile(hcl, []byte(workflowNamed("ci", "ci")), 0o644); err != nil {
		t.Fatal(err)
	}

	yaml := filepath.Join(tmp, "ci.yaml")
	if err := os.WriteFile(yaml, []byte("name: CI\non:\n  push: {}\njobs:\n  b:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	t.Run("parse", func(t *testing.T) {
		err := New().Parse(provider.ProviderOps{File: tmp, OutputDirectory: out})

		if !errors.Is(err, errFileIsADirectory) {
			t.Errorf("Parse(File: <dir>) = %v, want errFileIsADirectory", err)
		}
	})

	t.Run("unparse", func(t *testing.T) {
		err := New().Unparse(provider.ProviderOps{File: tmp, OutputDirectory: out})

		if !errors.Is(err, errFileIsADirectory) {
			t.Errorf("Unparse(File: <dir>) = %v, want errFileIsADirectory", err)
		}
	})

	t.Run("directory still works", func(t *testing.T) {
		if err := New().Parse(provider.ProviderOps{Directory: tmp, OutputDirectory: out}); err != nil {
			t.Errorf("Parse(Directory: <dir>) = %v, want the directory read", err)
		}
	})

	t.Run("file still works", func(t *testing.T) {
		if err := New().Parse(provider.ProviderOps{File: hcl, OutputDirectory: out}); err != nil {
			t.Errorf("Parse(File: <file>) = %v, want the file read", err)
		}
	})
}
