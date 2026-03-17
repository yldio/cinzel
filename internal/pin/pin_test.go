// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockResolver struct {
	shas map[string]string
}

func (m *mockResolver) ResolveTag(_ context.Context, owner, repo, tag string) (string, error) {
	key := fmt.Sprintf("%s/%s@%s", owner, repo, tag)

	if sha, ok := m.shas[key]; ok {
		return sha, nil
	}

	return "", fmt.Errorf("tag not found: %s", key)
}

func TestFindActionRefs(t *testing.T) {
	content := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "setup" {
  uses {
    action  = "actions/setup-go"
    version = "v5"
  }
}

step "pinned" {
  uses {
    action  = "actions/checkout"
    version = "de0fac2e4500dabe0009e67214ff5f5447ce83dd"
  }
}`

	refs := findActionRefs(content)

	if len(refs) != 3 {
		t.Fatalf("expected 3 refs, got %d", len(refs))
	}

	if refs[0].Action != "actions/checkout" || refs[0].Version != "v4" || !refs[0].IsTag {
		t.Errorf("ref[0]: got %+v", refs[0])
	}

	if refs[1].Action != "actions/setup-go" || refs[1].Version != "v5" || !refs[1].IsTag {
		t.Errorf("ref[1]: got %+v", refs[1])
	}

	if refs[2].IsTag {
		t.Error("ref[2] should not be a tag (it's a SHA)")
	}
}

func TestIsTag(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"v4", true},
		{"v1.2.3", true},
		{"v5.0", true},
		{"1.2.3", true},
		{"de0fac2e4500dabe0009e67214ff5f5447ce83dd", false},
		{"abc123", false},
		{"latest", false},
		{"main", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := isTag(tt.version); got != tt.want {
				t.Errorf("isTag(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestPinFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "checkout" {
  // actions/checkout v3
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "setup" {
  uses {
    action  = "actions/setup-go"
    version = "v5"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{
		shas: map[string]string{
			"actions/checkout@v4":  "abc123def456abc123def456abc123def456abc1",
			"actions/setup-go@v5": "def456abc123def456abc123def456abc123def4",
		},
	}

	var buf bytes.Buffer

	results, err := PinFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Action, r.Error)
		}

		if r.SHA == "" {
			t.Errorf("expected SHA for %s", r.Action)
		}
	}

	// Verify file was updated.
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(updated), `"v4"`) {
		t.Error("v4 tag should have been replaced")
	}

	if !strings.Contains(string(updated), `"abc123def456abc123def456abc123def456abc1"`) {
		t.Error("expected checkout SHA in output")
	}

	if !strings.Contains(string(updated), `"def456abc123def456abc123def456abc123def4"`) {
		t.Error("expected setup-go SHA in output")
	}

	// Verify output messages.
	output := buf.String()

	if !strings.Contains(output, "pinned actions/checkout@v4") {
		t.Error("expected pin message for checkout")
	}

	if !strings.Contains(output, "pinned actions/setup-go@v5") {
		t.Error("expected pin message for setup-go")
	}
}

func TestPinFileAlreadyPinned(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "de0fac2e4500dabe0009e67214ff5f5447ce83dd"
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

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if !results[0].WasAlready {
		t.Error("expected WasAlready to be true")
	}
}

func TestPinFileResolveFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "checkout" {
  uses {
    action  = "private/action"
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

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("expected error for unresolvable tag")
	}

	// File should remain unchanged.
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(updated), `"v1"`) {
		t.Error("version should remain as tag when resolution fails")
	}

	if !strings.Contains(buf.String(), "warning") {
		t.Error("expected warning in output")
	}
}

func TestPinFileAddsCommentWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "setup" {
  uses {
    action  = "actions/setup-go"
    version = "v5"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{
		shas: map[string]string{
			"actions/setup-go@v5": "def456abc123def456abc123def456abc123def4",
		},
	}

	var buf bytes.Buffer

	_, err := PinFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(updated), "// actions/setup-go v5") {
		t.Errorf("expected comment to be added\ngot:\n%s", string(updated))
	}

	// Verify indent: comment should have same indent as "uses {"
	if !strings.Contains(string(updated), "  // actions/setup-go v5\n  uses {") {
		t.Errorf("comment should have same indent as uses block\ngot:\n%s", string(updated))
	}
}

func TestPinFileUpdatesExistingComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "checkout" {
  // actions/checkout v3
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{
		shas: map[string]string{
			"actions/checkout@v4": "abc123def456abc123def456abc123def456abc1",
		},
	}

	var buf bytes.Buffer

	_, err := PinFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(updated), "// actions/checkout v4") {
		t.Errorf("expected comment to be updated to v4\ngot:\n%s", string(updated))
	}

	if strings.Contains(string(updated), "// actions/checkout v3") {
		t.Error("old comment v3 should have been replaced")
	}
}

func TestPinDirectory(t *testing.T) {
	dir := t.TempDir()

	file1 := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}`

	file2 := `step "setup" {
  uses {
    action  = "actions/setup-go"
    version = "v5"
  }
}`

	if err := os.WriteFile(filepath.Join(dir, "a.hcl"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "b.hcl"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}

	// Non-HCL file should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Docs"), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{
		shas: map[string]string{
			"actions/checkout@v4":  "sha1sha1sha1sha1sha1sha1sha1sha1sha1sha1",
			"actions/setup-go@v5": "sha2sha2sha2sha2sha2sha2sha2sha2sha2sha2",
		},
	}

	var buf bytes.Buffer

	results, err := PinDirectory(context.Background(), dir, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestPinFileDryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{
		shas: map[string]string{
			"actions/checkout@v4": "abc123def456abc123def456abc123def456abc1",
		},
	}

	var buf bytes.Buffer

	results, err := PinFile(context.Background(), path, resolver, &buf, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 || results[0].SHA == "" {
		t.Fatal("expected 1 resolved result")
	}

	// File should NOT be modified in dry-run mode.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(after), `"v4"`) {
		t.Error("dry-run should not modify the file")
	}
}

func TestPinDirectoryNoHCL(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Docs"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	_, err := PinDirectory(context.Background(), dir, &mockResolver{}, &buf, false)
	if err == nil {
		t.Error("expected error for directory with no HCL files")
	}
}

func TestCacheKey(t *testing.T) {
	key1 := cacheKey("actions", "checkout", "v4")
	key2 := cacheKey("actions", "checkout", "v5")
	key3 := cacheKey("actions", "checkout", "v4")

	if key1 == key2 {
		t.Error("different tags should produce different keys")
	}

	if key1 != key3 {
		t.Error("same inputs should produce same key")
	}
}
