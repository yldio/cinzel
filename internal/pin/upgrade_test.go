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

// A dry run has to find the upgrade and then decline to write it. Checking
// only that the file is unchanged does not say that: a run that resolved
// nothing leaves the file alone too, and so does an UpgradeFile that returns
// immediately. The upgrade it reports is what separates the two.
func TestUpgradeFileDryRun(t *testing.T) {
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

	resolver := &stubUpgrader{
		latestTags: map[string]string{"actions/checkout": "v5"},
		shas:       map[string]string{"actions/checkout@v5": "1111111111111111111111111111111111111111"},
	}

	var buf bytes.Buffer

	results, err := UpgradeFile(context.Background(), path, resolver, &buf, true)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected one action to be reported, got %d", len(results))
	}

	if results[0].NewTag != "v5" {
		t.Errorf("expected the dry run to report v5, got %q", results[0].NewTag)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(after) != content {
		t.Errorf("dry-run should not modify the file, got:\n%s", after)
	}
}

func TestUpgradeDirectoryNoHCL(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Docs"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	resolver := &stubUpgrader{}

	_, err := UpgradeDirectory(context.Background(), dir, resolver, &buf, false)
	if err == nil {
		t.Error("expected error for directory with no HCL files")
	}
}

type stubUpgrader struct {
	latestTags map[string]string
	shas       map[string]string
}

func (s *stubUpgrader) ResolveTag(_ context.Context, owner, repo, tag string) (string, error) {
	key := fmt.Sprintf("%s/%s@%s", owner, repo, tag)

	if sha, ok := s.shas[key]; ok {
		return sha, nil
	}

	return "", fmt.Errorf("tag not found: %s", key)
}

func (s *stubUpgrader) LatestTag(_ context.Context, owner, repo string) (string, error) {
	key := fmt.Sprintf("%s/%s", owner, repo)

	if tag, ok := s.latestTags[key]; ok {
		return tag, nil
	}

	return "", fmt.Errorf("no releases for %s", key)
}

func TestUpgradeInlineComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "aabbccddeeff00112233445566778899aabbccdd"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &stubUpgrader{
		latestTags: map[string]string{"actions/checkout": "v6"},
		shas:       map[string]string{"actions/checkout@v6": "b80b16730d25b9d3b6b2df7ca91e17d1ca6b9ef5"},
	}

	var buf bytes.Buffer

	_, err := UpgradeFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(updated), `"b80b16730d25b9d3b6b2df7ca91e17d1ca6b9ef5" # v6`) {
		t.Errorf("expected inline tag comment on version line\ngot:\n%s", string(updated))
	}
}
