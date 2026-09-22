// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A GitHub job or workflow keyword written with no value under it means the
// same as one left out, and parse drops a null written for one. Unparse wrote
// it all the same, so "if:" came out as "if = null" for cinzel's own parse to
// delete: the HCL from the first pass and the HCL from the second differed by
// a line that carried nothing.
//
// Two are worse than that. "defaults" and "strategy" are blocks in the schema,
// so "defaults = null" is HCL cinzel's own parse refuses to open, with
// `An argument named "defaults" is not expected here`.
func TestADroppedNullIsNotWritten(t *testing.T) {
	for _, tc := range []struct {
		name    string
		yaml    string
		notWant string
	}{
		{"a job block keyword", jobWithNull("defaults"), "defaults"},
		{"a job strategy", jobWithNull("strategy"), "strategy"},
		{"a job scalar keyword", jobWithNull("timeout-minutes"), "timeout_minutes"},
		{"a job condition", jobWithNull("if"), "if"},
		{"a job name", jobWithNull("name"), "name"},
		{"a job concurrency", jobWithNull("concurrency"), "concurrency"},
		{"a job container", jobWithNull("container"), "container"},
		{"a job environment", jobWithNull("environment"), "environment"},
		{"a job permissions", jobWithNull("permissions"), "permissions"},
		{"a job continue-on-error", jobWithNull("continue-on-error"), "continue_on_error"},
		{"a workflow defaults", workflowWithNull("defaults"), "defaults"},
		{"a workflow concurrency", workflowWithNull("concurrency"), "concurrency"},
		{"a workflow run-name", workflowWithNull("run-name"), "run_name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, _ := roundtripWorkflowYAML(t, tc.yaml)

			// hclwrite pads the attribute names of a block to a common width,
			// so the gap before "=" depends on the longest name beside it.
			flat := strings.Join(strings.Fields(hcl), " ")

			if strings.Contains(flat, tc.notWant+" = null") {
				t.Errorf("unparse wrote a null parse drops: %q\nHCL:\n%s", tc.notWant, hcl)
			}
		})
	}
}

// An env or with value written with no value under it is a name GitHub defines
// as empty, not a name it leaves undefined, so that null is the value and is
// kept. Unparse wrote it and parse then refused the file with "value must be
// set", which is unparse writing HCL its own parse cannot open.
func TestAnEnvOrWithNullIsKept(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			name: "a step env value",
			yaml: `on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
        env:
          V:
`,
		},
		{
			name: "a step with value",
			yaml: `on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          token:
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripWorkflowYAML(t, tc.yaml)

			// hclwrite pads the attribute names of a block to a common width,
			// so the gap before "=" depends on the longest name beside it.
			flat := strings.Join(strings.Fields(hcl), " ")

			if !strings.Contains(flat, "value = null") {
				t.Errorf("the null value was lost\nHCL:\n%s", hcl)
			}

			if !strings.Contains(back, "V:") && !strings.Contains(back, "token:") {
				t.Errorf("the name did not survive the trip back\nYAML:\n%s", back)
			}
		})
	}
}

// The HCL a second pass writes is the HCL the first one did. The first pass
// used to emit lines parse deleted, and the blank line the dropped keyword was
// separated by outlived it, so the two differed either way.
func TestHCLIsStableAcrossTwoPasses(t *testing.T) {
	const yml = `name: CI
on: push
permissions: {}
run-name:
defaults:
concurrency:
jobs:
  build:
    runs-on: ubuntu-latest
    defaults:
    strategy:
    timeout-minutes:
    if:
    name:
    concurrency:
    container:
    environment:
    continue-on-error:
    steps:
      - name: Test
        run: go test ./...
        env:
          V:
`

	first, back := roundtripWorkflowYAML(t, yml)
	second, _ := roundtripWorkflowYAML(t, back)

	if first != second {
		t.Errorf("HCL changed on the second pass\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// jobWithNull is a workflow whose build job sets keyword with nothing under it.
func jobWithNull(keyword string) string {
	return `name: CI
on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    ` + keyword + `:
    steps:
      - name: Test
        run: go test ./...
`
}

// workflowWithNull is a workflow setting keyword, at the top level, with
// nothing under it.
func workflowWithNull(keyword string) string {
	return `name: CI
on: push
permissions: {}
` + keyword + `:
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Test
        run: go test ./...
`
}

// roundtripWorkflowYAML unparses a workflow and parses the result back,
// returning the emitted HCL and the YAML it produced.
func roundtripWorkflowYAML(t *testing.T, yml string) (string, string) {
	t.Helper()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "w.yml")
	hclDir := filepath.Join(tmp, "hcl")
	backDir := filepath.Join(tmp, "yaml")

	if err := os.WriteFile(in, []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	hclPath := filepath.Join(hclDir, "w.hcl")
	hcl, err := os.ReadFile(hclPath)

	if err != nil {
		t.Fatalf("Unparse() wrote no HCL: %v", err)
	}

	if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: backDir}); err != nil {
		t.Fatalf("Parse() error = %v\nHCL:\n%s", err, hcl)
	}

	back, err := os.ReadFile(filepath.Join(backDir, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	return string(hcl), string(back)
}

// An action file goes through its own writer, which separated its sections the
// same way and left the same blank line behind when a keyword under "runs"
// wrote nothing.
func TestActionHCLIsStableAcrossTwoPasses(t *testing.T) {
	const yml = `name: My Action
description: does things
inputs:
  token:
    description: a token
    required:
    default:
runs:
  using: node20
  main: index.js
  pre:
  env:
    K:
`

	tmp := t.TempDir()
	in := filepath.Join(tmp, "in", "action.yml")

	if err := os.MkdirAll(filepath.Dir(in), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(in, []byte(yml), 0o600); err != nil {
		t.Fatal(err)
	}

	p := New()

	firstDir := filepath.Join(tmp, "hcl1")

	if err := p.Unparse(provider.ProviderOps{File: in, OutputDirectory: firstDir}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	hclPath := onlyFileUnder(t, firstDir)
	first := readFile(t, hclPath)

	backDir := filepath.Join(tmp, "yaml")

	if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: backDir}); err != nil {
		t.Fatalf("Parse() error = %v\nHCL:\n%s", err, first)
	}

	secondDir := filepath.Join(tmp, "hcl2")

	if err := p.Unparse(provider.ProviderOps{File: onlyFileUnder(t, backDir), OutputDirectory: secondDir}); err != nil {
		t.Fatalf("Unparse() error on the second pass = %v", err)
	}

	if second := readFile(t, onlyFileUnder(t, secondDir)); first != second {
		t.Errorf("HCL changed on the second pass\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	// hclwrite pads the attribute names of a block to a common width, so the
	// gap before "=" depends on the longest name beside it.
	flat := strings.Join(strings.Fields(first), " ")

	if !strings.Contains(flat, "value = null") {
		t.Errorf("the null env value was lost\nHCL:\n%s", first)
	}

	for _, notWant := range []string{"required", "default", "pre"} {
		if strings.Contains(flat, notWant+" = null") {
			t.Errorf("unparse wrote a null parse drops: %q\nHCL:\n%s", notWant, first)
		}
	}
}

// onlyFileUnder returns the one file written under dir. An action is named
// after the directory holding it, which a temporary directory decides, so the
// file is found rather than named.
func onlyFileUnder(t *testing.T, dir string) string {
	t.Helper()

	var found []string

	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			found = append(found, path)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if len(found) != 1 {
		t.Fatalf("want one file under %s, found %v", dir, found)
	}

	return found[0]
}
