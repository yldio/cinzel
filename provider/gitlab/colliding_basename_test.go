// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// The output name comes from the input's basename alone, so a recursive run
// over two directories each holding a ".gitlab-ci.yml" used to write one file
// twice and lose the first result without a word.
func TestCollidingBasenamesWriteDistinctFiles(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in")
	out := filepath.Join(tmp, "out")

	for _, dir := range []string{"a", "b"} {
		path := filepath.Join(in, dir)

		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}

		yml := "job-" + dir + ":\n  script:\n    - echo " + dir + "\n"

		if err := os.WriteFile(filepath.Join(path, ".gitlab-ci.yml"), []byte(yml), 0o600); err != nil {
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

	for _, want := range []string{"job-a", "job-b"} {
		if !strings.Contains(joined, want) {
			t.Errorf("want %q in the output, got:\n%s", want, joined)
		}
	}
}
