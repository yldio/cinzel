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

// An action's own comments went nowhere. The workflow writer carried them and
// the action writer did not: every top-level attribute went through
// writeAttributeAny, which takes none, so the note above the first key — the
// one describing the whole file — was read off the YAML and dropped. The
// blocks had the other half of it, with no body on the decoded struct for
// parse to read a block's comments back from.
func TestAnActionKeepsItsComments(t *testing.T) {
	const action = `# the file's own note
name: setup
# above description
description: sets up # beside description
inputs:
  # above the whole input
  token:
    # above the input description
    description: a token
outputs:
  # above the whole output
  result:
    description: the result
    value: done
runs:
  using: composite
  steps:
    - run: echo setup
      shell: bash
`

	tmp := t.TempDir()
	in := filepath.Join(tmp, "action.yml")

	if err := os.WriteFile(in, []byte(action), 0o600); err != nil {
		t.Fatal(err)
	}

	hclDir := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir}); err != nil {
		t.Fatalf("unparse: %v", err)
	}

	back := filepath.Join(tmp, "back")

	if err := New().Parse(provider.ProviderOps{Directory: hclDir, OutputDirectory: back}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	got := readOneFile(t, back)

	for _, want := range []string{
		"# the file's own note",
		"# above description",
		"# beside description",
		"# above the whole input",
		"# above the input description",
		"# above the whole output",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q:\n%s", want, got)
		}
	}
}

// readOneFile returns the contents of the single file written under dir, at
// whatever depth an action's own path put it.
func readOneFile(t *testing.T, dir string) string {
	t.Helper()

	var found string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		found = string(content)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if found == "" {
		t.Fatalf("no file written under %s", dir)
	}

	return found
}
