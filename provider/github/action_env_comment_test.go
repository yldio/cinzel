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

// An action's "runs.env" lost its comments in both directions: parse wrote the
// value with no comments attached, and unparse called the name-value writer
// with a nil tree. A comment written above an env entry was gone from the YAML,
// and one written in the YAML was gone from the HCL, with nothing said either
// time.
func TestActionRunsEnvKeepsItsComments(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")
	back := filepath.Join(dir, "back")
	path := filepath.Join(dir, "in.hcl")

	hcl := stepBlock("hi", "") +
		"action \"alpha\" {\n" +
		"  filename = \"alpha\"\n" +
		"  name = \"A\"\n" +
		"  description = \"d\"\n" +
		"  runs {\n" +
		"    using = \"composite\"\n" +
		"    // what it holds\n" +
		"    env {\n      name = \"TOKEN_NAME\"\n      value = \"v\"\n    }\n" +
		"    steps = [step.hi]\n" +
		"  }\n}\n"

	if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	yml, err := os.ReadFile(filepath.Join(out, "alpha", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(yml), "# what it holds\n    TOKEN_NAME:") {
		t.Errorf("parse dropped the env comment:\n%s", yml)
	}

	if err := New().Unparse(provider.ProviderOps{Directory: out, Recursive: true, OutputDirectory: back}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(back, "alpha.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(got), "# what it holds\n    env {") {
		t.Errorf("unparse dropped the env comment:\n%s", got)
	}
}
