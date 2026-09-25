// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

// The directory walk compared the extension exactly, where the single-file
// branch checks none at all and ListFilesWithExtensions folds the ones it is
// given. macOS and Windows open "build.HCL" as the same file either way, so a
// file read when named on the command line was skipped when found in a
// directory, and skipped without a word: the gitlab pipeline came out missing
// that file's jobs at exit 0, and github read the YAML it had already
// generated for it as stale and deleted it.
func TestAnHCLFileIsFoundWhateverItsExtensionCase(t *testing.T) {
	for _, name := range []string{"upper.HCL", "mixed.Hcl", "lower.hcl"} {
		t.Run(name, func(t *testing.T) {
			tmp := t.TempDir()

			path := filepath.Join(tmp, name)
			if err := os.WriteFile(path, []byte("x = 1\n"), 0o644); err != nil {
				t.Fatal(err)
			}

			body, sources, err := ParseHCLInput(tmp, false)
			if err != nil {
				t.Fatalf("ParseHCLInput() = %v, want the file read", err)
			}

			if body == nil {
				t.Fatal("ParseHCLInput() returned no body")
			}

			if _, found := sources[path]; !found {
				t.Errorf("sources have no entry for %s, got %v", path, keys(sources))
			}
		})
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))

	for k := range m {
		out = append(out, k)
	}

	return out
}
