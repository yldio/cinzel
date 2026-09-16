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

// writeYAML puts a workflow in a directory and returns that directory.
func writeYAML(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()

	for name, content := range files {
		path := filepath.Join(dir, name)

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// roundtrip converts a directory of YAML to HCL and straight back, returning
// the directory holding the regenerated YAML.
func roundtrip(t *testing.T, yamlDir string) string {
	t.Helper()

	hclDir := t.TempDir()
	backDir := t.TempDir()

	if err := New().Unparse(provider.ProviderOps{Directory: yamlDir, OutputDirectory: hclDir, Recursive: true}); err != nil {
		t.Fatalf("unparse: %v", err)
	}

	if err := New().Parse(provider.ProviderOps{Directory: hclDir, OutputDirectory: backDir}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	return backDir
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

const oneStepNoIDWorkflow = `name: w
on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`

// A step written without an "id" came back carrying one, because parse gives a
// step with no id the name of its block label and unparse had no way to say
// the source never had one. The YAML gained a line nobody wrote.
func TestAStepWithNoIDDoesNotGainOne(t *testing.T) {
	back := roundtrip(t, writeYAML(t, map[string]string{"w.yaml": oneStepNoIDWorkflow}))

	got := readFile(t, filepath.Join(back, "w.yaml"))

	if strings.Contains(got, "id:") {
		t.Errorf("roundtrip invented a step id:\n%s", got)
	}
}

// An "id" the author did write has to survive, which is the other half of the
// same rule: ignore_id must not be written for a step that has one.
func TestAStepWithAnIDKeepsIt(t *testing.T) {
	const withID = `name: w
on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - id: mine
        run: echo hi
`

	back := roundtrip(t, writeYAML(t, map[string]string{"w.yaml": withID}))

	got := readFile(t, filepath.Join(back, "w.yaml"))

	if !strings.Contains(got, "id: mine") {
		t.Errorf("roundtrip dropped the step id:\n%s", got)
	}
}

// Two workflows each holding a step of the same name produced two step blocks
// with the same label. Parse merges every file in a directory into one body,
// so it refused to read back output cinzel had just written.
func TestTheSameStepNameInTwoFilesStillParses(t *testing.T) {
	const first = `name: a
on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: mise setup
        run: mise install
`

	const second = `name: b
on: push
permissions: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: mise setup
        run: mise install --verbose
`

	back := roundtrip(t, writeYAML(t, map[string]string{"a.yaml": first, "b.yaml": second}))

	for _, name := range []string{"a.yaml", "b.yaml"} {
		if _, err := os.Stat(filepath.Join(back, name)); err != nil {
			t.Errorf("%s was not written back: %v", name, err)
		}
	}
}

// An action's steps share the label space with every workflow in the run, so
// the collision reaches across the two kinds of document too.
func TestAnActionAndAWorkflowDoNotShareAStepLabel(t *testing.T) {
	const action = `name: my action
description: a thing
runs:
  using: composite
  steps:
    - run: echo hi
      shell: bash
`

	dir := writeYAML(t, map[string]string{
		"w.yaml":              oneStepNoIDWorkflow,
		"myaction/action.yml": action,
	})

	back := roundtrip(t, dir)

	if _, err := os.Stat(filepath.Join(back, "w.yaml")); err != nil {
		t.Errorf("workflow was not written back: %v", err)
	}

	if _, err := os.Stat(filepath.Join(back, "myaction", "action.yml")); err != nil {
		t.Errorf("action was not written back: %v", err)
	}
}
