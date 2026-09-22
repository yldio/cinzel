// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSubdirectoryActionIsPinned covers an action living in a subdirectory of
// its repository. "github/codeql-action/init" is the "init" directory of the
// repository "github/codeql-action", and the tag sits on the repository, so
// everything past the second segment names a path inside it rather than a
// repository of its own.
func TestSubdirectoryActionIsPinned(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "init" {
  uses {
    action  = "github/codeql-action/init"
    version = "v3"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{
		shas: map[string]string{
			"github/codeql-action@v3": "abc123def456abc123def456abc123def456abc1",
		},
	}

	var buf bytes.Buffer

	results, err := PinFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(updated), `"abc123def456abc123def456abc123def456abc1" # v3`) {
		t.Errorf("expected the subdirectory action to be pinned, got:\n%s", updated)
	}
}

// TestSubdirectoryActionIsUpgraded is the same shape on the upgrade side, which
// asks for the latest release of the repository holding the subdirectory.
func TestSubdirectoryActionIsUpgraded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "init" {
  uses {
    action  = "github/codeql-action/init"
    version = "v2"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &stubUpgrader{
		latestTags: map[string]string{"github/codeql-action": "v3"},
		shas:       map[string]string{"github/codeql-action@v3": "abc123def456abc123def456abc123def456abc1"},
	}

	var buf bytes.Buffer

	results, err := UpgradeFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(updated), `"abc123def456abc123def456abc123def456abc1" # v3`) {
		t.Errorf("expected the subdirectory action to be upgraded, got:\n%s", updated)
	}
}

// TestLocalActionIsNotResolved covers the control: an action written as a path
// into this repository names no GitHub repository, so nothing is asked of the
// API and nothing is rewritten.
func TestLocalActionIsNotResolved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "local" {
  uses {
    action  = "./.github/actions/build"
    version = "v1"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{shas: map[string]string{}}

	var buf bytes.Buffer

	results, err := PinFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 || results[0].Error == nil {
		t.Fatalf("expected a local action to be reported, got %+v", results)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(updated) != content {
		t.Errorf("expected the file to be left alone, got:\n%s", updated)
	}
}
