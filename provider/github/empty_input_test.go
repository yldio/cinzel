// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// parseHCLFile runs Parse over HCL written inline, returning the error and the
// directory the output was asked for, so a test can check both the error and
// whether anything was written.
func parseHCLFile(t *testing.T, hcl string) (error, string) {
	t.Helper()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "input.hcl")

	if err := os.WriteFile(in, []byte(hcl), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	return New().Parse(provider.ProviderOps{File: in, OutputDirectory: out}), out
}

func TestEmptyInputIsRejected(t *testing.T) {
	for name, hcl := range map[string]string{
		"empty file":      "",
		"only a comment":  "// nothing here\n",
		"only whitespace": "\n\n\n",
	} {
		t.Run(name, func(t *testing.T) {
			err, out := parseHCLFile(t, hcl)

			if !errors.Is(err, errNoDefinitions) {
				t.Fatalf("expected errNoDefinitions, got %v", err)
			}

			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Error("nothing should have been written")
			}
		})
	}
}

// TestEmptyInputLeavesGeneratedFilesAlone is the reason the guard exists. An
// empty input used to be written as a "{}" document, which then became the only
// current output, so pruning deleted every other generated file in the
// directory.
func TestEmptyInputLeavesGeneratedFilesAlone(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	out := filepath.Join(tmp, "out")

	for _, dir := range []string{src, out} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(src, "empty.hcl"), []byte("// nothing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	existing := filepath.Join(out, "ci.yaml")
	content := []byte(generatedByLine + "\n# cinzel-provider: github\nname: CI\n")

	if err := os.WriteFile(existing, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{Directory: src, OutputDirectory: out}); !errors.Is(err, errNoDefinitions) {
		t.Fatalf("expected errNoDefinitions, got %v", err)
	}

	if _, err := os.Stat(existing); err != nil {
		t.Fatalf("a generated file was deleted by an empty parse: %v", err)
	}
}

// TestStepOnlyInputStillParses guards the fix: a real step-only file is the
// same code path and must keep working.
func TestStepOnlyInputStillParses(t *testing.T) {
	err, out := parseHCLFile(t, `step "test" {
  run = "go test ./..."
}
`)
	if err != nil {
		t.Fatalf("a real step-only input should parse: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(out, "input.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(got), "go test ./...") {
		t.Errorf("step body missing from output:\n%s", got)
	}
}

// unparseDir runs Unparse over a directory of YAML files written inline.
func unparseDir(t *testing.T, files map[string]string) (error, string) {
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

const realWorkflow = `name: Real
on: push
jobs:
  a:
    runs-on: ubuntu-latest
    steps:
      - run: echo a
`

// A file holding no document used to be rejected by the step decoder, which
// aborted the whole run before the real files were reached.
func TestEmptyFileDoesNotAbortDirectoryUnparse(t *testing.T) {
	for name, body := range map[string]string{
		"empty file":     "",
		"only a comment": "# nothing here\n",
	} {
		t.Run(name, func(t *testing.T) {
			err, out := unparseDir(t, map[string]string{
				"real.yml":  realWorkflow,
				"stray.yml": body,
			})
			if err != nil {
				t.Fatalf("a stray %s should be skipped, got %v", name, err)
			}

			if _, statErr := os.Stat(filepath.Join(out, "real.hcl")); statErr != nil {
				t.Errorf("the real workflow was not converted: %v", statErr)
			}

			if _, statErr := os.Stat(filepath.Join(out, "stray.hcl")); !os.IsNotExist(statErr) {
				t.Error("the stray file should not have produced HCL")
			}
		})
	}
}

// TestMalformedFileStillErrors guards the fix: skipping an empty document must
// not turn real syntax errors into silence.
func TestMalformedFileStillErrors(t *testing.T) {
	err, _ := unparseDir(t, map[string]string{"bad.yml": "name: X\n  bad indent: [\n"})
	if err == nil {
		t.Fatal("malformed YAML should still be an error")
	}
}
