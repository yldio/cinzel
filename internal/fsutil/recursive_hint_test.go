// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// A directory holding HCL only below its top level is what a mirrored
// conversion produces, so cinzel's own output read back without --recursive
// lands here. "No HCL files found" is true of the top level and says nothing
// about the rest, which reads as "your files are missing".
func TestHCLOnlyBelowTheTopLevelNamesTheFix(t *testing.T) {
	tmp := t.TempDir()

	nested := filepath.Join(tmp, "actions")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(nested, "setup.hcl"), []byte("step \"s\" {\n  run = \"echo hi\"\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := ParseHCLInput(tmp, false)

	if !errors.Is(err, ErrNoHCLFilesAtTopLevel) {
		t.Fatalf("ParseHCLInput(.., false) = %v, want ErrNoHCLFilesAtTopLevel", err)
	}

	// And the same input read the way the message says to read it.
	if _, _, err := ParseHCLInput(tmp, true); err != nil {
		t.Errorf("ParseHCLInput(.., true) = %v, want the nested file read", err)
	}
}

// A directory with nothing in it anywhere gets the plain message: there is
// no --recursive to suggest, because recursing would find nothing either.
func TestAnEmptyDirectoryGetsThePlainMessage(t *testing.T) {
	tmp := t.TempDir()

	_, _, err := ParseHCLInput(tmp, false)

	if !errors.Is(err, ErrNoHCLFiles) {
		t.Fatalf("ParseHCLInput() = %v, want ErrNoHCLFiles", err)
	}

	if errors.Is(err, ErrNoHCLFilesAtTopLevel) {
		t.Error("an empty directory was told to try --recursive")
	}
}

// Nested files that are not HCL are not a reason to suggest recursing.
func TestNestedNonHCLDoesNotSuggestRecursive(t *testing.T) {
	tmp := t.TempDir()

	nested := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(nested, "notes.txt"), []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := ParseHCLInput(tmp, false)

	if errors.Is(err, ErrNoHCLFilesAtTopLevel) {
		t.Errorf("a directory holding no HCL at all was told to try --recursive: %v", err)
	}
}

// Already reading recursively, so there is nothing to suggest.
func TestARecursiveReadGetsThePlainMessage(t *testing.T) {
	tmp := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmp, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}

	_, _, err := ParseHCLInput(tmp, true)

	if errors.Is(err, ErrNoHCLFilesAtTopLevel) {
		t.Errorf("a recursive read was told to try --recursive: %v", err)
	}
}
