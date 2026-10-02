// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A pipeline cinzel converts, and one it recognises as a pipeline and refuses:
// a "needs" entry naming no job at all.
const (
	goodPipeline = `stages:
  - build
build:
  stage: build
  script:
    - make
`

	badPipeline = `stages:
  - build
build:
  stage: build
  needs:
    - ""
  script:
    - make
`

	alsoBadPipeline = `stages:
  - test
test:
  stage: test
  needs:
    - ""
  script:
    - make test
`
)

// unparseGitLabDir runs Unparse over a directory of YAML files written inline,
// returning the error and the directory the output was asked for, so a test
// can check both the error and whether anything was written.
func unparseGitLabDir(t *testing.T, files map[string]string) (error, string) {
	t.Helper()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in")

	if err := os.MkdirAll(in, 0o755); err != nil {
		t.Fatal(err)
	}

	for name, body := range files {
		if err := os.WriteFile(filepath.Join(in, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out := filepath.Join(tmp, "out")

	return New().Unparse(provider.ProviderOps{Directory: in, OutputDirectory: out}), out
}

// A file that fails used to abort the run where it stood, so the files
// converted before it were already written and the ones after it were never
// read. Which half survived was decided by where the failing file sorted.
func TestAFailedPipelineLeavesNothingWritten(t *testing.T) {
	for name, files := range map[string]map[string]string{
		// The bug. "a-good" is converted before "z-bad" is read.
		"the good file sorts first": {"a-good.yml": goodPipeline, "z-bad.yml": badPipeline},
		// The control: this order wrote nothing even before the fix.
		"the bad file sorts first": {"a-bad.yml": badPipeline, "z-good.yml": goodPipeline},
	} {
		t.Run(name, func(t *testing.T) {
			err, out := unparseGitLabDir(t, files)

			if err == nil {
				t.Fatal("want the run to fail, got no error")
			}

			if entries, _ := os.ReadDir(out); len(entries) != 0 {
				t.Errorf("want nothing written, got %d entries", len(entries))
			}
		})
	}
}

// Every failure is reported, not the first one the run met.
func TestEveryFailedPipelineIsReported(t *testing.T) {
	err, _ := unparseGitLabDir(t, map[string]string{
		"a-bad.yml": badPipeline,
		"z-bad.yml": alsoBadPipeline,
	})

	if err == nil {
		t.Fatal("want the run to fail, got no error")
	}

	for _, want := range []string{"a-bad.yml", "z-bad.yml"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want %q named in the error, got: %v", want, err)
		}
	}
}

// A file holding no pipeline is passed over and the run goes on, the same as
// before. Only a file that is a pipeline and then fails stops the run.
func TestANonPipelineStillLetsTheRunWrite(t *testing.T) {
	err, out := unparseGitLabDir(t, map[string]string{
		"a-other.yml": "name: not a pipeline\n",
		"z-good.yml":  goodPipeline,
	})
	if err != nil {
		t.Fatalf("want the run to succeed, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(out, "z-good.hcl")); statErr != nil {
		t.Errorf("want the good file written: %v", statErr)
	}
}

// "found" was set before the conversion rather than after it, so a run whose
// only pipeline failed had already recorded that it found one. Nothing read
// that until the failure stopped being returned on the spot; now that the run
// reaches the end, a directory whose every pipeline failed must report the
// failures rather than reporting that nothing was found.
func TestAFailedPipelineIsNotReportedAsNoDefinitions(t *testing.T) {
	err, _ := unparseGitLabDir(t, map[string]string{"only.yml": badPipeline})

	if err == nil {
		t.Fatal("want the run to fail, got no error")
	}

	if errors.Is(err, errNoDefinitions) {
		t.Errorf("want the failure reported, got %v", err)
	}

	if !errors.Is(err, errNeedsJobEmpty) {
		t.Errorf("want %v, got %v", errNeedsJobEmpty, err)
	}
}
