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

const unknownKeyWorkflow = `name: ci
on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    bogus-job-key: x
    steps:
      - run: make
        bogus-step-key: y
`

// Refusing is the default, and the file is left alone. A dropped key is gone
// from the HCL and so from the YAML the next parse writes over the source, so
// losing one is a choice the run has to make for itself.
func TestAnUnknownKeyIsRefusedByDefault(t *testing.T) {
	err := unparseErr(t, unknownKeyWorkflow)

	if err == nil {
		t.Fatal("expected the unknown key to be refused")
	}

	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("the error does not name the key: %v", err)
	}
}

// With the flag the keys are dropped and each one is named, at the line it
// was written on. The strict decode this stands in for stops at the first key
// it does not know, so a file with several had to be run once per key.
func TestIgnoreUnknownDropsEveryKeyAndNamesIt(t *testing.T) {
	hcl, warnings := unparseWithOpts(t, unknownKeyWorkflow, provider.ProviderOps{IgnoreUnknown: true})

	for _, want := range []string{
		":7: jobs.build.bogus-job-key",
		":10: jobs.build.steps[0].bogus-step-key",
	} {
		if !strings.Contains(warnings, want) {
			t.Errorf("want a report at %q, got:\n%s", want, warnings)
		}
	}

	if strings.Contains(hcl, "bogus") {
		t.Errorf("a dropped key reached the HCL:\n%s", hcl)
	}

	// What the schema does declare is untouched.
	for _, want := range []string{`name = "ci"`, `runners = "ubuntu-latest"`, `run = "make"`} {
		if !strings.Contains(hcl, want) {
			t.Errorf("want %q in:\n%s", want, hcl)
		}
	}
}

// An action is the other document type and takes the same path.
func TestIgnoreUnknownOnAnAction(t *testing.T) {
	const action = `name: setup
description: sets up
typo: oops
runs:
  using: composite
  steps:
    - run: echo
      shell: bash
`

	hcl, warnings := unparseWithOpts(t, action, provider.ProviderOps{IgnoreUnknown: true})

	if !strings.Contains(warnings, ":3: typo") {
		t.Errorf("the dropped key was not reported at its line:\n%s", warnings)
	}

	if strings.Contains(hcl, "typo") {
		t.Errorf("the dropped key reached the HCL:\n%s", hcl)
	}
}

// A document the schema fully declares reports nothing, with the flag set or
// not: the flag changes what happens to a key the schema does not know, and
// nothing else.
func TestIgnoreUnknownIsSilentOnACleanFile(t *testing.T) {
	const clean = `name: ci
on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make
`

	_, warnings := unparseWithOpts(t, clean, provider.ProviderOps{IgnoreUnknown: true})

	if strings.Contains(warnings, "dropping") {
		t.Errorf("nothing was dropped, but something was reported:\n%s", warnings)
	}
}

// unparseWithOpts converts the YAML with the given options, returning the HCL
// and whatever the run wrote to stderr.
func unparseWithOpts(t *testing.T, yaml string, opts provider.ProviderOps) (string, string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.yaml")
	output := filepath.Join(dir, "out")

	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	opts.File = path
	opts.OutputDirectory = output

	warnings := captureGitHubStderr(t, func() {
		if err := New().Unparse(opts); err != nil {
			t.Fatalf("Unparse() = %v", err)
		}
	})

	entries, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}

	var hcl strings.Builder

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := os.ReadFile(filepath.Join(output, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}

		hcl.Write(content)
	}

	return hcl.String(), warnings
}
