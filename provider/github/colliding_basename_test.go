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

// The output name comes from the input's basename alone, so a recursive run
// over two directories each holding a "ci.yaml" used to write one file twice
// and lose the first result without a word.
func TestCollidingBasenamesWriteDistinctFiles(t *testing.T) {
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

	// Both inputs have to be present. Whichever file each landed in, the two
	// names together are what proves nothing was overwritten.
	var joined string

	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(out, e.Name()))
		if err != nil {
			t.Fatal(err)
		}

		joined += string(b)
	}

	for _, want := range []string{"from-a", "from-b"} {
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
