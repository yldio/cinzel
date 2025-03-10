// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestListFilesWithExtensions(t *testing.T) {
	tmp := t.TempDir()

	rootYAML := filepath.Join(tmp, "a.yaml")
	nestedDir := filepath.Join(tmp, "nested")
	nestedYML := filepath.Join(nestedDir, "b.yml")
	other := filepath.Join(tmp, "c.txt")

	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, f := range []string{rootYAML, nestedYML, other} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("non-recursive", func(t *testing.T) {
		files, err := ListFilesWithExtensions(tmp, false, ".yaml", ".yml")
		if err != nil {
			t.Fatal(err)
		}

		if len(files) != 1 || files[0] != rootYAML {
			t.Fatalf("unexpected files: %#v", files)
		}
	})

	t.Run("recursive", func(t *testing.T) {
		files, err := ListFilesWithExtensions(tmp, true, ".yaml", ".yml")
		if err != nil {
			t.Fatal(err)
		}

		if len(files) != 2 {
			t.Fatalf("expected 2 yaml files, got %#v", files)
		}
	})
}

func TestListFilesWithExtensionsFileValidation(t *testing.T) {
	tmp := t.TempDir()
	bad := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(bad, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ListFilesWithExtensions(bad, false, ".yaml", ".yml")
	if err == nil {
		t.Fatal("expected extension validation error")
	}
}

func TestWriteFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "a", "b", "c.txt")

	if err := WriteFile(target, []byte("hello")); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "hello" {
		t.Fatalf("expected hello, got %q", string(b))
	}
}

func TestParseHCLInputNoFiles(t *testing.T) {
	tmp := t.TempDir()
	_, err := ParseHCLInput(tmp, false)
	if !errors.Is(err, ErrNoHCLFiles) {
		t.Fatalf("expected ErrNoHCLFiles, got %v", err)
	}
}
