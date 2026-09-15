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

// unparseYAMLString runs Unparse over YAML written inline and returns the
// error together with the directory the output was asked for.
func unparseYAMLString(t *testing.T, yaml string) (error, string) {
	t.Helper()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "wf.yaml")

	if err := os.WriteFile(in, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	return New().Unparse(provider.ProviderOps{File: in, OutputDirectory: out}), out
}

const firstDoc = `name: first
on:
  push:
jobs:
  a:
    runs-on: ubuntu-latest
    steps:
      - run: echo a
`

// GitHub reads one document per file, and a file holding two used to be cut
// short at the first with nothing said and a zero exit.
func TestMultipleDocumentsAreRejected(t *testing.T) {
	err, out := unparseYAMLString(t, firstDoc+`---
name: second
on:
  push:
jobs:
  b:
    runs-on: ubuntu-latest
    steps:
      - run: echo b
`)

	if !errors.Is(err, errMultipleDocuments) {
		t.Fatalf("want errMultipleDocuments, got %v", err)
	}

	if entries, _ := os.ReadDir(out); len(entries) != 0 {
		t.Errorf("want nothing written, got %d entries", len(entries))
	}
}

// A stray "---" at either end is a marker, not a second document.
func TestStrayDocumentMarkersAreFine(t *testing.T) {
	for name, yaml := range map[string]string{
		"leading":  "---\n" + firstDoc,
		"trailing": firstDoc + "---\n",
		"both":     "---\n" + firstDoc + "---\n",
	} {
		t.Run(name, func(t *testing.T) {
			err, out := unparseYAMLString(t, yaml)
			if err != nil {
				t.Fatalf("unparse: %v", err)
			}

			got, err := os.ReadFile(filepath.Join(out, "wf.hcl"))
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(got), `name = "first"`) {
				t.Errorf("want the workflow kept, got:\n%s", got)
			}
		})
	}
}

// A single document keeps working, which is the case every other test covers
// but the one this change could most easily break.
func TestSingleDocumentStillUnparses(t *testing.T) {
	err, out := unparseYAMLString(t, firstDoc)
	if err != nil {
		t.Fatalf("unparse: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "wf.hcl")); err != nil {
		t.Errorf("want output written: %v", err)
	}
}
