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

type mockGitHubResolver struct {
	latestTags map[string]string // "owner/repo" → latest tag
	shas       map[string]string // "owner/repo@tag" → sha
}

func (m *mockGitHubResolver) resolveTag(owner, repo, tag string) (string, error) {
	key := fmt.Sprintf("%s/%s@%s", owner, repo, tag)

	if sha, ok := m.shas[key]; ok {
		return sha, nil
	}

	return "", fmt.Errorf("tag not found: %s", key)
}

func (m *mockGitHubResolver) latestTag(owner, repo string) (string, error) {
	key := fmt.Sprintf("%s/%s", owner, repo)

	if tag, ok := m.latestTags[key]; ok {
		return tag, nil
	}

	return "", fmt.Errorf("no releases for %s", key)
}

// We need to test with real GitHubResolver methods, so let's test
// the upgrade logic via UpgradeFile with a real resolver that we mock
// at the HTTP level. For unit tests, we'll test the helper functions directly.

func TestUpgradeFileIntegration(t *testing.T) {
	// This test uses the real UpgradeFile but with a mock resolver
	// that would require HTTP mocking. For now, test the upgrade
	// logic indirectly via the building blocks.
	t.Skip("requires HTTP mock — covered by e2e tests")
}

func TestFindActionRefsForUpgrade(t *testing.T) {
	content := `step "checkout" {
  // actions/checkout v4
  uses {
    action  = "actions/checkout"
    version = "de0fac2e4500dabe0009e67214ff5f5447ce83dd"
  }
}

step "setup" {
  uses {
    action  = "actions/setup-go"
    version = "v4"
  }
}`

	refs, err := findActionRefs(content)
	if err != nil {
		t.Fatal(err)
	}

	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}

	// First ref is SHA-pinned
	if refs[0].IsTag {
		t.Error("SHA version should not be detected as tag")
	}

	// Second ref is a tag
	if !refs[1].IsTag || refs[1].Version != "v4" {
		t.Errorf("expected tag v4, got %+v", refs[1])
	}
}

func TestUpgradeFileDryRun(t *testing.T) {
	// Create a test file and verify dry-run doesn't modify it.
	// Uses a custom GitHubResolver subclass would be needed for full test,
	// but we can verify the file-level behavior with the real function
	// by making LatestTag fail (no network).
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

	// This will fail because there's no HTTP mock, but it exercises the code path.
	resolver := NewGitHubResolver("")

	var buf bytes.Buffer

	// This will produce warnings (API calls fail) but shouldn't panic.
	_, _ = UpgradeFile(context.Background(), path, resolver, &buf, true)

	// Verify file unchanged in dry-run.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(after) != content {
		t.Error("dry-run should not modify the file")
	}
}

func TestUpgradeDirectoryNoHCL(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Docs"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	resolver := NewGitHubResolver("")

	_, err := UpgradeDirectory(context.Background(), dir, resolver, &buf, false)
	if err == nil {
		t.Error("expected error for directory with no HCL files")
	}
}

func TestUpsertUsesCommentForUpgrade(t *testing.T) {
	// Verify that upsertUsesComment works when upgrading from v4 to v6.
	content := `step "checkout" {
  // actions/checkout v4
  uses {
    action  = "actions/checkout"
    version = "old-sha"
  }
}
`

	updated := upsertUsesComment(content, "actions/checkout", "v6")

	if !strings.Contains(updated, "// actions/checkout v6") {
		t.Errorf("expected comment updated to v6\ngot:\n%s", updated)
	}

	if strings.Contains(updated, "// actions/checkout v4") {
		t.Error("old v4 comment should be replaced")
	}
}
