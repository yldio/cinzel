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

// A matrix axis is named by whoever wrote the HCL and referenced by that name
// in "${{ matrix.X }}", which is copied through as written. Renaming the axis
// on the way to YAML and not the reference left the reference resolving against
// an axis that no longer existed, which actionlint reads as `property
// "go_version" is not defined in object type {go-version: ...}`.
//
// The two ways of writing an axis also disagreed: an attribute was renamed, a
// "variable" block was not, so the same axis came out spelled two ways.
func TestParseKeepsMatrixAxisNames(t *testing.T) {
	const src = `step "setup" {
  name = "Setup"
  run  = "go version $${{ matrix.go_version }} $${{ matrix.node_version }}"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  strategy {
    fail_fast = false

    matrix {
      go_version = ["1.24", "1.25"]

      variable {
        name  = "node_version"
        value = ["20"]
      }

      include = [{
        go_version = "1.26"
      }]
    }
  }

  steps = [step.setup]
}

workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`

	dir := t.TempDir()
	input := filepath.Join(dir, "ci.hcl")

	if err := os.WriteFile(input, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()

	if err := New().Parse(provider.ProviderOps{File: input, OutputDirectory: outDir}); err != nil {
		t.Fatal(err)
	}

	gotPath, err := singleYAMLFileInDir(outDir)
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatal(err)
	}

	yaml := string(got)

	for _, axis := range []string{"go_version:", "node_version:"} {
		if !strings.Contains(yaml, axis) {
			t.Errorf("axis %q is missing from the matrix\ngot:\n%s", axis, yaml)
		}
	}

	for _, renamed := range []string{"go-version:", "node-version:"} {
		if strings.Contains(yaml, renamed) {
			t.Errorf("axis was renamed to %q, which no ${{ matrix.* }} reference names\ngot:\n%s", renamed, yaml)
		}
	}

	// An include entry carries axis names too, and was already left alone.
	if !strings.Contains(yaml, "go_version: \"1.26\"") {
		t.Errorf("the include entry lost its axis name\ngot:\n%s", yaml)
	}

	// Everything outside the matrix is still spelled the way YAML wants it.
	for _, key := range []string{"runs-on:", "fail-fast:"} {
		if !strings.Contains(yaml, key) {
			t.Errorf("%q is missing: the rename should still apply outside the matrix\ngot:\n%s", key, yaml)
		}
	}
}
