// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// workflowNamed returns a workflow HCL file writing to filename.
func workflowNamed(id string, filename string) string {
	return "step \"" + id + "_echo\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"" + id + "_build\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n" +
		"  steps = [step." + id + "_echo]\n}\n\n" +
		"workflow \"" + id + "\" {\n  filename = \"" + filename + "\"\n" +
		"  on \"push\" {}\n  jobs = [job." + id + "_build]\n}\n"
}

// TestNarrowedParseKeepsAnotherFilesOutput is the bug this ownership marker
// exists for. Parsing one HCL file used to delete the output of every other
// one, because the prune walked the output directory and deleted whatever the
// current run had not just written — and a run given one file writes one file.
func TestNarrowedParseKeepsAnotherFilesOutput(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	out := filepath.Join(tmp, "out")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, f := range []struct{ name, id string }{{"ci.hcl", "ci"}, {"cd.hcl", "cd"}} {
		if err := os.WriteFile(filepath.Join(src, f.name), []byte(workflowNamed(f.id, f.id)), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if err := New().Parse(provider.ProviderOps{Directory: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"ci.yaml", "cd.yaml"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("the directory run did not write %s: %v", name, err)
		}
	}

	if err := New().Parse(provider.ProviderOps{File: filepath.Join(src, "cd.hcl"), OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(out, "ci.yaml")); err != nil {
		t.Fatalf("parsing cd.hcl deleted ci.hcl's output, which it never read: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "cd.yaml")); err != nil {
		t.Fatalf("expected the parsed file's own output, got %v", err)
	}
}

// TestParseRecordsTheDeclaringFile pins that the recorded source names the file
// the block was written in rather than the directory the run was pointed at.
// Every file in a directory is merged into one body before the decode, and a
// source that survives that merge is what makes a narrowed run decidable.
func TestParseRecordsTheDeclaringFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	out := filepath.Join(tmp, "out")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, f := range []struct{ name, id string }{{"ci.hcl", "ci"}, {"cd.hcl", "cd"}} {
		if err := os.WriteFile(filepath.Join(src, f.name), []byte(workflowNamed(f.id, f.id)), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if err := New().Parse(provider.ProviderOps{Directory: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	for _, want := range []struct{ output, source string }{{"ci.yaml", "ci.hcl"}, {"cd.yaml", "cd.hcl"}} {
		content, err := os.ReadFile(filepath.Join(out, want.output))
		if err != nil {
			t.Fatal(err)
		}

		line := "# cinzel-source: "

		if !strings.Contains(string(content), line) {
			t.Fatalf("%s records no source:\n%s", want.output, content)
		}

		if !strings.Contains(string(content), want.source+"\n") {
			t.Errorf("%s should name %s, got:\n%s", want.output, want.source, content)
		}
	}
}

// TestNarrowedParseKeepsAnotherFilesAction is the same narrowing over actions,
// which are written into a directory of their own rather than beside the
// workflows and so take a different path through the prune.
func TestNarrowedParseKeepsAnotherFilesAction(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	out := filepath.Join(tmp, "out")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}

	action := func(id string) string {
		return "action \"" + id + "\" {\n  filename = \"" + id + "\"\n  name = \"" + id + "\"\n\n" +
			"  runs {\n    using = \"node20\"\n    main  = \"index.js\"\n  }\n}\n"
	}

	for _, id := range []string{"first", "second"} {
		if err := os.WriteFile(filepath.Join(src, id+".hcl"), []byte(action(id)), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if err := New().Parse(provider.ProviderOps{Directory: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: filepath.Join(src, "second.hcl"), OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(out, "first", "action.yml")); err != nil {
		t.Fatalf("parsing second.hcl deleted first.hcl's action, which it never read: %v", err)
	}
}

// TestGeneratedMarkersDoNotReachTheHCL guards the strip in
// WithoutGeneratedMarker from the other end: the source line names an HCL file,
// so a line left in the YAML would be written back into the file it names, one
// more comment on every roundtrip.
func TestGeneratedMarkersDoNotReachTheHCL(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	out := filepath.Join(tmp, "out")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(src, "ci.hcl"), []byte(workflowNamed("ci", "ci")), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{Directory: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	back := filepath.Join(tmp, "back")

	if err := New().Unparse(provider.ProviderOps{Directory: out, OutputDirectory: back}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(back)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(back, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}

		for _, marker := range []string{"generated-by: cinzel", "cinzel-provider:", "cinzel-source:"} {
			if strings.Contains(string(content), marker) {
				t.Errorf("%s carries %q into the HCL:\n%s", entry.Name(), marker, content)
			}
		}
	}
}
