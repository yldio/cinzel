// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A workflow cinzel converts, and one it recognises as a workflow and refuses:
// "needs" names a job the file does not declare.
const (
	goodWorkflow = `name: Good
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make
`

	badWorkflow = `name: Bad
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    needs: [nope]
    steps:
      - run: make
`

	alsoBadWorkflow = `name: Also bad
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    needs: [missing]
    steps:
      - run: make
`
)

// A file that fails used to abort the run where it stood, so the files
// converted before it were already written and the ones after it were never
// read. Which half survived was decided by where the failing file sorted, and
// a half-converted directory reads as a finished one: the next step is usually
// deleting the YAML, and the workflows never reached go with it.
func TestAFailedFileLeavesNothingWritten(t *testing.T) {
	for name, files := range map[string]map[string]string{
		// The bug. "a-good" is converted before "z-bad" is read.
		"the good file sorts first": {"a-good.yml": goodWorkflow, "z-bad.yml": badWorkflow},
		// The control: this order wrote nothing even before the fix, so on
		// its own it would prove the run stops rather than that it buffers.
		"the bad file sorts first": {"a-bad.yml": badWorkflow, "z-good.yml": goodWorkflow},
	} {
		t.Run(name, func(t *testing.T) {
			err, out := unparseDir(t, files)

			if err == nil {
				t.Fatal("want the run to fail, got no error")
			}

			if entries, _ := os.ReadDir(out); len(entries) != 0 {
				t.Errorf("want nothing written, got %d entries", len(entries))
			}
		})
	}
}

// Every failure is reported, not the first one the run met. A directory of
// five broken workflows otherwise takes five runs to read through, one key at
// a time.
func TestEveryFailedFileIsReported(t *testing.T) {
	err, _ := unparseDir(t, map[string]string{
		"a-bad.yml": badWorkflow,
		"z-bad.yml": alsoBadWorkflow,
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

// The other outcome, which must stay as it is. A file that is not a workflow,
// an action or a set of steps is warned about and skipped, because cinzel
// never had it and nothing is lost by passing over it. Only a file cinzel
// recognises and then fails on stops the run.
func TestASkippedFileStillLetsTheRunWrite(t *testing.T) {
	err, out := unparseDir(t, map[string]string{
		"a-skipped.yml": "version: 2\nupdates:\n  - package-ecosystem: gomod\n",
		"z-good.yml":    goodWorkflow,
	})
	if err != nil {
		t.Fatalf("want the run to succeed, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(out, "z-good.hcl")); statErr != nil {
		t.Errorf("want the good file written: %v", statErr)
	}

	if _, statErr := os.Stat(filepath.Join(out, "a-skipped.hcl")); !os.IsNotExist(statErr) {
		t.Error("want the skipped file not written")
	}
}

// A dry run printed every file up to the one that failed, which is the same
// partial result on stdout rather than on disk.
func TestADryRunWithAFailedFilePrintsNothing(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in")

	if err := os.MkdirAll(in, 0o755); err != nil {
		t.Fatal(err)
	}

	for name, body := range map[string]string{"a-good.yml": goodWorkflow, "z-bad.yml": badWorkflow} {
		if err := os.WriteFile(filepath.Join(in, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var unparseErr error

	printed := captureUnparseStdout(t, func() {
		unparseErr = New().Unparse(provider.ProviderOps{
			Directory:       in,
			OutputDirectory: filepath.Join(tmp, "out"),
			DryRun:          true,
		})
	})

	if unparseErr == nil {
		t.Fatal("want the run to fail, got no error")
	}

	if strings.Contains(printed, "a-good") {
		t.Errorf("want nothing printed, got:\n%s", printed)
	}
}

// captureUnparseStdout returns whatever fn wrote to os.Stdout. A dry run
// prints there rather than writing, so there is nothing else to assert on.
func captureUnparseStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	original := os.Stdout
	os.Stdout = w

	defer func() { os.Stdout = original }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return string(out)
}
