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

// unparsedActionHCL converts one action document and returns the HCL.
func unparsedActionHCL(t *testing.T, action string) string {
	t.Helper()

	tmp := t.TempDir()
	dir := filepath.Join(tmp, "actions", "thing")

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "action.yml"), []byte(action), 0o600); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "hcl")

	err := New().Unparse(provider.ProviderOps{
		Root:            tmp,
		Directory:       filepath.Join("actions", "thing"),
		OutputDirectory: out,
	})
	if err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(out, "thing.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// Every step closes with a blank line, and the action block went straight on
// after the last one: the "}" and the "action" that followed it sat on
// consecutive lines, so the file read as one run of blocks rather than a set
// of them.
func TestAnActionIsSeparatedFromTheStepsAboveIt(t *testing.T) {
	hcl := unparsedActionHCL(t, "name: Thing\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: One\n      run: a\n      shell: bash\n    - name: Two\n      run: b\n      shell: bash\n")

	lines := strings.Split(hcl, "\n")

	for i, line := range lines {
		if !strings.HasPrefix(line, "action ") {
			continue
		}

		if i == 0 {
			t.Fatalf("the action is the first line, so there are no steps above it:\n%s", hcl)
		}

		if strings.TrimSpace(lines[i-1]) != "" {
			t.Errorf("line %d is %q, want a blank line before the action:\n%s", i, lines[i-1], hcl)
		}

		return
	}

	t.Fatalf("no action block in the output:\n%s", hcl)
}

// Nothing above it to separate from, so no blank line to add. A file opening
// on an empty line is its own untidiness.
func TestAnActionWithNoStepsStartsAtTheTop(t *testing.T) {
	hcl := unparsedActionHCL(t, "name: Thing\ndescription: d\nruns:\n  using: docker\n  image: alpine:3.20\n")

	if !strings.HasPrefix(hcl, "action ") {
		t.Errorf("want the file to open on the action block, got:\n%s", hcl)
	}
}

// The spacing is cosmetic, so what it must not do is change what the file
// means.
func TestTheSeparatedActionStillRoundTrips(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "actions", "thing")

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	action := "name: Thing\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: One\n      run: a\n      shell: bash\n"

	if err := os.WriteFile(filepath.Join(dir, "action.yml"), []byte(action), 0o600); err != nil {
		t.Fatal(err)
	}

	hcl := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{Root: tmp, Directory: filepath.Join("actions", "thing"), OutputDirectory: hcl}); err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	back := filepath.Join(tmp, "back")

	if err := New().Parse(provider.ProviderOps{Root: back, Directory: hcl}); err != nil {
		t.Fatalf("Parse() = %v", err)
	}

	if _, err := os.Stat(filepath.Join(back, "actions", "thing", "action.yml")); err != nil {
		t.Errorf("want the action back: %v", err)
	}
}
