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

const pipelineYAML = "build:\n  stage: build\n  script:\n    - echo hi\n"

// writeYAML puts one file in a fresh directory and returns both paths.
func writeYAML(t *testing.T, name, content string) (in, out string) {
	t.Helper()

	dir := t.TempDir()
	in = filepath.Join(dir, "in")
	out = filepath.Join(dir, "out")

	if err := os.MkdirAll(in, 0o750); err != nil {
		t.Fatal(err)
	}

	if name != "" {
		if err := os.WriteFile(filepath.Join(in, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return in, out
}

// Reporting success having written nothing says the input was converted when
// it was not. Pointing at the wrong directory lands here, and so does
// forgetting --recursive with the pipeline a level down.
func TestUnparseRefusesInputWithNoYAML(t *testing.T) {
	in, out := writeYAML(t, "", "")

	err := New().Unparse(provider.ProviderOps{Directory: in, OutputDirectory: out})
	if !errors.Is(err, errNoYAMLFiles) {
		t.Fatalf("want errNoYAMLFiles, got %v", err)
	}
}

// A pipeline one directory down, with --recursive left off, is the same
// silence reached by a plainer mistake.
func TestUnparseRefusesWhenRecursiveIsMissing(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in")
	out := filepath.Join(dir, "out")

	deep := filepath.Join(in, "deep")
	if err := os.MkdirAll(deep, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(deep, ".gitlab-ci.yml"), []byte(pipelineYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{Directory: in, OutputDirectory: out}); !errors.Is(err, errNoYAMLFiles) {
		t.Fatalf("want errNoYAMLFiles, got %v", err)
	}

	// The same input with the flag on is the control: the refusal is about
	// finding nothing, not about the directory being nested.
	if err := New().Unparse(provider.ProviderOps{Directory: in, Recursive: true, OutputDirectory: out}); err != nil {
		t.Fatalf("want the nested pipeline converted, got %v", err)
	}
}

// YAML that holds no pipeline is read and then skipped, so the run ends
// having written nothing. The same silence hid the same mistake.
func TestUnparseRefusesYAMLWithNoPipeline(t *testing.T) {
	in, out := writeYAML(t, "notes.yml", "just: a mapping\nnothing: to do with ci\n")

	err := New().Unparse(provider.ProviderOps{Directory: in, OutputDirectory: out})
	if !errors.Is(err, errNoDefinitions) {
		t.Fatalf("want errNoDefinitions, got %v", err)
	}
}

// A pipeline still converts, and a dry run still succeeds: it found the
// pipeline and skipped only the write it was told to skip.
func TestUnparseStillConvertsAPipeline(t *testing.T) {
	for _, dry := range []bool{false, true} {
		in, out := writeYAML(t, ".gitlab-ci.yml", pipelineYAML)

		if err := New().Unparse(provider.ProviderOps{Directory: in, DryRun: dry, OutputDirectory: out}); err != nil {
			t.Fatalf("dry-run=%v: want the pipeline converted, got %v", dry, err)
		}

		if dry {
			continue
		}

		if _, err := os.Stat(filepath.Join(out, ".gitlab-ci.hcl")); err != nil {
			t.Errorf("want the output written, got %v", err)
		}
	}
}

// A YAML file that will not parse reports the parser's complaint on its own,
// naming a line and column in a file it does not name. A directory run reads
// every ".yaml" in the tree, most of which are not pipelines at all, so the
// one line the reader has to go on is which file it was.
func TestUnparseNamesTheFileThatWillNotParse(t *testing.T) {
	in, out := writeYAML(t, "broken.yaml", "key: [unclosed\n")

	err := New().Unparse(provider.ProviderOps{Directory: in, OutputDirectory: out})
	if err == nil {
		t.Fatal("want an error")
	}

	if !strings.Contains(err.Error(), "broken.yaml") {
		t.Errorf("error does not name the file: %v", err)
	}
}
