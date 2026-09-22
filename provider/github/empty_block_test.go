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

// parseEmptyBlockHCL converts src and returns whatever the parse refused, with the
// YAML it wrote when it refused nothing.
func parseEmptyBlockHCL(t *testing.T, src string) (error, string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "w.hcl")

	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out")
	err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: out})

	written, _ := os.ReadFile(filepath.Join(out, "w.yaml"))

	return err, string(written)
}

// A block with an empty body became an empty map, which goes out as a bare
// "key:" with a null under it. GitHub rejects all of these, and the runs-on
// form cinzel's own unparse refuses too, so a parse exited 0 on a file nothing
// downstream could read.
func TestABlockThatSetsNothingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
		key  string
	}{
		{name: "a job runs_on", src: workflowWithExtra("", "\n  runs_on {}\n"), key: "runs_on"},
		{name: "a job strategy", src: workflowWithExtra("", "\n  strategy {}\n"), key: "strategy"},
		{name: "a job container", src: workflowWithExtra("", "\n  container {}\n"), key: "container"},
		{name: "a job environment", src: workflowWithExtra("", "\n  environment {}\n"), key: "environment"},
		{name: "a job concurrency", src: workflowWithExtra("", "\n  concurrency {}\n"), key: "concurrency"},
		{name: "a job defaults", src: workflowWithExtra("", "\n  defaults {}\n"), key: "defaults"},
		{name: "a job service", src: workflowWithExtra("", "\n  service \"db\" {}\n"), key: "service"},
		{name: "a workflow defaults", src: workflowWithExtra("\n  defaults {}\n", ""), key: "defaults"},
		{name: "a workflow concurrency", src: workflowWithExtra("\n  concurrency {}\n", ""), key: "concurrency"},
		{name: "a strategy matrix", src: workflowWithExtra("", "\n  strategy {\n    matrix {}\n  }\n"), key: "matrix"},
		{name: "a defaults run", src: workflowWithExtra("\n  defaults {\n    run {}\n  }\n", ""), key: "run"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, written := parseEmptyBlockHCL(t, tc.src)

			if err == nil {
				t.Fatalf("expected an error, got:\n%s", written)
			}

			if !errors.Is(err, errEmptyBlock) {
				t.Errorf("error is not the empty-block one: %v", err)
			}

			if !strings.Contains(err.Error(), tc.key) {
				t.Errorf("error does not name the block, got: %v", err)
			}
		})
	}
}

// An empty permissions block is the one that stays: an absent or null
// permissions field makes GitHub inherit the default token permissions rather
// than granting none, so writing "{}" is the only way to deny them all.
func TestAnEmptyPermissionsBlockIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
	}{
		{name: "on the workflow", src: workflowWithExtra("\n  permissions {}\n", "")},
		{name: "on the job", src: workflowWithExtra("", "\n  permissions {}\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, written := parseEmptyBlockHCL(t, tc.src)

			if err != nil {
				t.Fatalf("expected the block to be written, got %v", err)
			}

			if !strings.Contains(written, "permissions: {}") {
				t.Errorf("lost the explicit empty permissions:\n%s", written)
			}
		})
	}
}

// A block carrying anything at all is a block GitHub can read, and has to keep
// going through.
func TestABlockThatSetsSomethingIsKept(t *testing.T) {
	src := workflowWithExtra("", "\n  strategy {\n    fail_fast = false\n  }\n\n  container {\n    image = \"node:18\"\n  }\n")

	err, written := parseEmptyBlockHCL(t, src)

	if err != nil {
		t.Fatalf("expected the blocks to be written, got %v", err)
	}

	for _, want := range []string{"fail-fast: false", "image: \"node:18\""} {
		if !strings.Contains(written, want) {
			t.Errorf("lost %q:\n%s", want, written)
		}
	}
}
