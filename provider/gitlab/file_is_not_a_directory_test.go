// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// --file names a file, the same way it does in the github provider. Both
// resolveInputPath copies returned the path untouched and left the directory
// case to the stat further down, so --file walked a directory --directory
// exists for.
func TestFileFlagRefusesADirectory(t *testing.T) {
	tmp := t.TempDir()

	yaml := filepath.Join(tmp, ".gitlab-ci.yml")
	if err := os.WriteFile(yaml, []byte("build:\n  script:\n    - make\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	t.Run("unparse", func(t *testing.T) {
		err := New().Unparse(provider.ProviderOps{File: tmp, OutputDirectory: out})

		if !errors.Is(err, errFileIsADirectory) {
			t.Errorf("Unparse(File: <dir>) = %v, want errFileIsADirectory", err)
		}
	})

	t.Run("directory still works", func(t *testing.T) {
		if err := New().Unparse(provider.ProviderOps{Directory: tmp, OutputDirectory: out}); err != nil {
			t.Errorf("Unparse(Directory: <dir>) = %v, want the directory read", err)
		}
	})

	t.Run("file still works", func(t *testing.T) {
		if err := New().Unparse(provider.ProviderOps{File: yaml, OutputDirectory: out}); err != nil {
			t.Errorf("Unparse(File: <file>) = %v, want the file read", err)
		}
	})
}
