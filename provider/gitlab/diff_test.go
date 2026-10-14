// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/provider"
)

const diffPipeline = `job "test" {
  script = ["go test ./..."]
  stage  = "test"
}
`

// A pipeline is one file and nothing here prunes, so a diff is that file
// against what is committed: no stale set, and never a report of a file the
// provider does not own.
func TestDiffReportsWhatParseWouldChange(t *testing.T) {
	setUp := func(t *testing.T) (hclPath string, outDir string) {
		t.Helper()

		tmp := t.TempDir()
		hclPath = filepath.Join(tmp, "pipeline.hcl")
		outDir = filepath.Join(tmp, "out")

		if err := os.WriteFile(hclPath, []byte(diffPipeline), 0o644); err != nil {
			t.Fatal(err)
		}

		if err := New().Parse(provider.ProviderOps{File: hclPath, OutputDirectory: outDir}); err != nil {
			t.Fatal(err)
		}

		return hclPath, outDir
	}

	t.Run("a pipeline parse just wrote is identical", func(t *testing.T) {
		hclPath, outDir := setUp(t)

		results, err := New().Diff(provider.ProviderOps{File: hclPath, OutputDirectory: outDir})
		if err != nil {
			t.Fatal(err)
		}

		if provider.Changed(results) {
			t.Errorf("Diff() = %v, want nothing to do", results)
		}
	})

	t.Run("an edited pipeline differs", func(t *testing.T) {
		hclPath, outDir := setUp(t)

		if err := os.WriteFile(filepath.Join(outDir, ".gitlab-ci.yml"), []byte("edited: by hand\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := New().Diff(provider.ProviderOps{File: hclPath, OutputDirectory: outDir})
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 1 || results[0].Status != provider.DiffDiffers {
			t.Errorf("Diff() = %v, want the pipeline reported as differing", results)
		}
	})

	// Nothing here prunes, so a marked file the provider did not plan is not
	// its to call stale. Asking anyway would report every generated file in
	// the directory.
	t.Run("another generated file is not called stale", func(t *testing.T) {
		hclPath, outDir := setUp(t)

		beside := filepath.Join(outDir, "other.yaml")

		if err := os.WriteFile(beside,
			[]byte("# generated-by: cinzel\n# cinzel-provider: gitlab\nname: other\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := New().Diff(provider.ProviderOps{File: hclPath, OutputDirectory: outDir})
		if err != nil {
			t.Fatal(err)
		}

		if provider.Changed(results) {
			t.Errorf("Diff() = %v, want the file beside the pipeline left alone", results)
		}

		if _, err := os.Stat(beside); err != nil {
			t.Errorf("a reporting call removed %s: %v", beside, err)
		}
	})
}
