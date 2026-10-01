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

// A step carrying neither "run" nor "uses" is something parse writes, so
// unparse has to read it back.
//
// Refusing it here read as a rule about what GitHub will run, and a step-only
// file is not something GitHub reads: it is cinzel's own library of step
// definitions, and a step in it may be a name and an env and nothing else.
// Narrowing the step-only path to documents that run something refused that
// file, so parse wrote HCL that unparse then exited 1 on.
//
// The mixed case is the sharper one. The check rejected the whole document on
// the first value that failed it, so one such step beside a working one lost
// both.
func TestAStepThatRunsNothingSurvivesARoundtrip(t *testing.T) {
	hcl := `step "documented" {
  name = "Just a name"

  env {
    name  = "FOO"
    value = "bar"
  }
}

step "real" {
  run = "make"
}
`

	tmp := t.TempDir()
	hclFile := filepath.Join(tmp, "steps.hcl")

	if err := os.WriteFile(hclFile, []byte(hcl), 0o644); err != nil {
		t.Fatal(err)
	}

	yamlDir := filepath.Join(tmp, "yaml")

	if err := New().Parse(provider.ProviderOps{File: hclFile, OutputDirectory: yamlDir}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	backDir := filepath.Join(tmp, "back")

	if err := New().Unparse(provider.ProviderOps{Directory: yamlDir, OutputDirectory: backDir}); err != nil {
		t.Fatalf("unparse of what parse just wrote: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(backDir, "steps.hcl"))
	if err != nil {
		t.Fatalf("no HCL came back: %v", err)
	}

	for _, want := range []string{`step "documented"`, "Just a name", "FOO", `step "real"`, "make"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("want %q in the HCL that came back:\n%s", want, got)
		}
	}
}
