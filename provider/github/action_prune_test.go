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

// An action carrying a path lands outside the output directory, and the prune
// only ever walked that one. Renaming such an action left the old action.yml
// on disk: marked as cinzel's, no longer written by anything, and still
// runnable by GitHub.
func TestRenamedActionWithAPathIsPruned(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "in.hcl")
	out := filepath.Join(tmp, "out")

	write := func(t *testing.T, name string) {
		t.Helper()

		if err := os.WriteFile(src, []byte(actionWithPathHCL("act", name, "actions-a")), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
			t.Fatal(err)
		}
	}

	write(t, "setup")
	write(t, "init")

	if _, err := os.Stat(filepath.Join(out, "actions-a", "setup", "action.yml")); !os.IsNotExist(err) {
		t.Errorf("the renamed-away action is still there, stat err=%v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "actions-a", "init", "action.yml")); err != nil {
		t.Errorf("want the current action kept: %v", err)
	}
}

// The prune now reaches directories it never used to, so what stops it there
// is the marker and nothing else. A file cinzel did not write is not cinzel's
// to remove, whatever it is called and wherever it sits.
func TestThePruneLeavesHandWrittenFilesBesideAnAction(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "in.hcl")
	out := filepath.Join(tmp, "out")

	if err := os.WriteFile(src, []byte(actionWithPathHCL("act", "setup", "actions-a")), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	handmade := filepath.Join(out, "actions-a", "handmade", "action.yml")
	loose := filepath.Join(out, "actions-a", "loose.yaml")

	if err := os.MkdirAll(filepath.Dir(handmade), 0o750); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{handmade, loose} {
		if err := os.WriteFile(path, []byte("name: mine\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{handmade, loose} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("the prune removed a file it did not write: %s: %v", path, err)
		}
	}
}

// A diff that cannot see outside the output directory reports green while an
// action has drifted, which is worse than no check at all.
func TestDiffSeesAnActionOutsideTheOutputDirectory(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "in.hcl")
	out := filepath.Join(tmp, "out")

	if err := os.WriteFile(src, []byte(actionWithPathHCL("act", "setup", "actions-a")), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	written := filepath.Join(out, "actions-a", "setup", "action.yml")

	content, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(written, []byte(strings.Replace(string(content), "composite", "node20", 1)), 0o600); err != nil {
		t.Fatal(err)
	}

	results, err := New().Diff(provider.ProviderOps{File: src, OutputDirectory: out})
	if err != nil {
		t.Fatal(err)
	}

	if !provider.Changed(results) {
		t.Errorf("an edited action outside the output directory read as clean: %v", results)
	}
}

// The walk recurses, so an output directory containing an action directory
// covers it. Asking for both reported one stale file twice: harmless to a
// prune, which removes it once, and wrong in a diff, which counted it.
func TestAStaleFileIsReportedOnce(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "in.hcl")

	if err := os.WriteFile(src, []byte(actionWithPathHCL("act", "setup", "actions-a")), 0o600); err != nil {
		t.Fatal(err)
	}

	// The output directory is the parent of the action directory, so the two
	// prune targets overlap.
	if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: tmp}); err != nil {
		t.Fatal(err)
	}

	orphan := filepath.Join(tmp, "actions-a", "orphan.yaml")
	marked := "# generated-by: cinzel\n# cinzel-provider: github\nname: orphan\n"

	if err := os.WriteFile(orphan, []byte(marked), 0o600); err != nil {
		t.Fatal(err)
	}

	results, err := New().Diff(provider.ProviderOps{File: src, OutputDirectory: tmp})
	if err != nil {
		t.Fatal(err)
	}

	seen := 0

	for _, result := range results {
		if result.Status == provider.DiffStale {
			seen++
		}
	}

	if seen != 1 {
		t.Errorf("the stale file was reported %d times, want once: %v", seen, results)
	}
}
