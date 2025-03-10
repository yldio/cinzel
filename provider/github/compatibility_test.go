// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

func TestProviderWiringSmoke(t *testing.T) {
	tmpDir := t.TempDir()
	hclIn := filepath.Join(tmpDir, "in.hcl")
	parseDir := filepath.Join(tmpDir, "parse")
	unparseDir := filepath.Join(tmpDir, "unparse")

	hcl := `step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "wf" {
  filename = "smoke"
  on "push" {}
  jobs = [job.build]
}
`

	if err := os.WriteFile(hclIn, []byte(hcl), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()
	if err := p.Parse(provider.ProviderOps{File: hclIn, OutputDirectory: parseDir}); err != nil {
		t.Fatal(err)
	}

	yamlOut := filepath.Join(parseDir, "smoke.yaml")
	yamlBytes, err := os.ReadFile(yamlOut)
	if err != nil {
		t.Fatal(err)
	}

	yamlOutStr := string(yamlBytes)
	if !strings.Contains(yamlOutStr, "on:") || !strings.Contains(yamlOutStr, "jobs:") {
		t.Fatalf("expected workflow yaml output, got:\n%s", yamlOutStr)
	}

	if err := p.Unparse(provider.ProviderOps{File: yamlOut, OutputDirectory: unparseDir}); err != nil {
		t.Fatal(err)
	}

	hclOut, err := os.ReadFile(filepath.Join(unparseDir, "smoke.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	hclOutStr := string(hclOut)
	if !strings.Contains(hclOutStr, `workflow "smoke"`) || !strings.Contains(hclOutStr, `job "build"`) {
		t.Fatalf("expected workflow and job blocks in hcl output, got:\n%s", hclOutStr)
	}
}
