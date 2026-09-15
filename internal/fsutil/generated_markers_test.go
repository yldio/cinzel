// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrependGeneratedMarker(t *testing.T) {
	content := []byte("name: CI\n")
	output := string(PrependGeneratedMarker(content, "github"))

	if !strings.Contains(output, "# generated-by: cinzel") {
		t.Fatalf("expected generated marker header, got: %q", output)
	}

	if !strings.Contains(output, "# cinzel-provider: github") {
		t.Fatalf("expected provider marker header, got: %q", output)
	}
}

func TestHasGeneratedMarker(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("matches provider", func(t *testing.T) {
		path := filepath.Join(tmpDir, "owned.yaml")
		content := "# generated-by: cinzel\n# cinzel-provider: github\nname: ci\n"

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		ok, err := HasGeneratedMarker(path, "github")
		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			t.Fatal("expected marker ownership to match provider")
		}
	})

	t.Run("does not match other provider", func(t *testing.T) {
		path := filepath.Join(tmpDir, "other.yaml")
		content := "# generated-by: cinzel\n# cinzel-provider: gitlab\nname: ci\n"

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		ok, err := HasGeneratedMarker(path, "github")
		if err != nil {
			t.Fatal(err)
		}

		if ok {
			t.Fatal("expected marker ownership mismatch for different provider")
		}
	})
}

func TestPruneStaleGeneratedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	currentPath := filepath.Join(tmpDir, "current.yaml")
	stalePath := filepath.Join(tmpDir, "stale.yaml")
	manualPath := filepath.Join(tmpDir, "manual.yaml")

	currentContent := "# generated-by: cinzel\n# cinzel-provider: github\nname: current\n"
	staleContent := "# generated-by: cinzel\n# cinzel-provider: github\nname: stale\n"
	manualContent := "name: manual\n"

	if err := os.WriteFile(currentPath, []byte(currentContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(stalePath, []byte(staleContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(manualPath, []byte(manualContent), 0o644); err != nil {
		t.Fatal(err)
	}

	currentOutputs := map[string]struct{}{filepath.Clean(currentPath): {}}
	if err := PruneStaleGeneratedYAML(tmpDir, currentOutputs, "github"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("expected stale provider-owned file removed, stat err=%v", err)
	}

	if _, err := os.Stat(manualPath); err != nil {
		t.Fatalf("expected manual file preserved, stat err=%v", err)
	}

	if _, err := os.Stat(currentPath); err != nil {
		t.Fatalf("expected current file preserved, stat err=%v", err)
	}
}

// An output can sit in a subdirectory, because a filename may name one. The
// prune read only the top level, so a file left there was never reached and
// stayed beside the current one for good.
func TestPruneReachesASubdirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "sub")

	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	owned := "# generated-by: cinzel\n# cinzel-provider: github\nname: x\n"

	files := map[string]string{
		filepath.Join(nested, "stale.yaml"):   owned,
		filepath.Join(nested, "current.yaml"): owned,
		filepath.Join(nested, "manual.yaml"):  "name: manual\n",
		// Another provider's output is not this one's to delete.
		filepath.Join(nested, "other.yaml"): "# generated-by: cinzel\n# cinzel-provider: gitlab\nname: x\n",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	current := map[string]struct{}{filepath.Clean(filepath.Join(nested, "current.yaml")): {}}

	if err := PruneStaleGeneratedYAML(dir, current, "github"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(nested, "stale.yaml")); !os.IsNotExist(err) {
		t.Errorf("want the nested stale file removed, stat err=%v", err)
	}

	for _, keep := range []string{"current.yaml", "manual.yaml", "other.yaml"} {
		if _, err := os.Stat(filepath.Join(nested, keep)); err != nil {
			t.Errorf("want %s kept, got %v", keep, err)
		}
	}
}

// An output directory that does not exist yet is the first run, not an error.
func TestPruneToleratesAMissingDirectory(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "not-there")

	if err := PruneStaleGeneratedYAML(missing, map[string]struct{}{}, "github"); err != nil {
		t.Errorf("want no error for a missing directory, got %v", err)
	}
}
