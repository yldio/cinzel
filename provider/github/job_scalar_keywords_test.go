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

// parseHCLToYAML runs Parse over HCL written inline and returns the single
// workflow file it produced.
func parseHCLToYAML(t *testing.T, hcl string) string {
	t.Helper()

	tmp := t.TempDir()
	in := filepath.Join(tmp, "wf.hcl")

	if err := os.WriteFile(in, []byte(hcl), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "out")

	if err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(out, "wf.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	return string(got)
}

// jobHCL wraps a job body in the smallest workflow that carries it.
func jobHCL(body string) string {
	return `
step "echo" {
  run = "echo a"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

` + body + `
  steps = [step.echo]
}

workflow "wf" {
  filename = "wf"

  on "push" {}

  jobs = [job.build]
}
`
}

// A job's "concurrency", "container" and "environment" each accept a plain
// string in GitHub's schema. Unparse writes that string as an attribute, so
// Parse has to read it back as one instead of demanding a block.
func TestJobScalarKeywordsParse(t *testing.T) {
	for name, tc := range map[string]struct{ attr, want string }{
		"concurrency": {`  concurrency = "ci-group"`, "concurrency: ci-group"},
		"container":   {`  container = "node:18"`, `container: "node:18"`},
		"environment": {`  environment = "production"`, "environment: production"},
	} {
		t.Run(name, func(t *testing.T) {
			got := parseHCLToYAML(t, jobHCL(tc.attr))

			if !strings.Contains(got, tc.want) {
				t.Errorf("want %q in output, got:\n%s", tc.want, got)
			}
		})
	}
}

// The block form of the same three keywords has to keep working: the attribute
// is an addition, not a replacement.
func TestJobScalarKeywordsStillAcceptBlocks(t *testing.T) {
	got := parseHCLToYAML(t, jobHCL(`
  concurrency {
    group              = "ci"
    cancel_in_progress = true
  }

  container {
    image = "node:18"
  }

  environment {
    name = "prod"
    url  = "https://example.com"
  }
`))

	for _, want := range []string{"group: ci", "cancel-in-progress: true", `image: "node:18"`, "name: prod"} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in output, got:\n%s", want, got)
		}
	}
}

// The roundtrip the bug actually broke: YAML holding the scalar form has to
// survive Unparse followed by Parse.
func TestJobScalarKeywordsRoundtrip(t *testing.T) {
	const yaml = `name: wf
on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
    concurrency: ci-group
    container: "node:18"
    environment: production
    steps:
      - run: make
`

	tmp := t.TempDir()
	in := filepath.Join(tmp, "wf.yaml")

	if err := os.WriteFile(in, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	hclDir := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir}); err != nil {
		t.Fatalf("unparse: %v", err)
	}

	hclBytes, err := os.ReadFile(filepath.Join(hclDir, "wf.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	got := parseHCLToYAML(t, string(hclBytes))

	for _, want := range []string{"concurrency: ci-group", `container: "node:18"`, "environment: production"} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q after roundtrip, got:\n%s", want, got)
		}
	}
}
