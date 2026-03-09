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

func TestParseValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		errContains string
	}{
		{
			name: "workflow missing on",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  jobs = [job.build]
}
`,
			errContains: "at least one trigger",
		},
		{
			name: "unknown workflow attribute",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  unknown = "x"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "unknown attribute 'unknown' in workflow",
		},
		{
			name: "unknown job block",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  mystery {}
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "unknown block 'mystery' in job",
		},
		{
			name: "workflow missing jobs",
			content: `workflow "ci" {
  filename = "ci"
  on "push" {}
}
`,
			errContains: "at least one job",
		},
		{
			name: "job with with but no uses",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  with {
    name = "a"
    value = "b"
  }
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "only valid when 'uses'",
		},
		{
			name: "job missing runs_on",
			content: `step "s" { run = "echo hi" }
job "build" {
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "'runs_on' is required",
		},
		{
			name: "job uses with steps",
			content: `step "s" { run = "echo hi" }
job "build" {
  uses = "org/repo/.github/workflows/reusable.yaml@v1"
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "cannot be defined together",
		},
		{
			name: "depends_on missing job",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  depends_on = [job.missing]
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "cannot find needed job",
		},
		{
			name: "duplicate depends_on",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  steps = [step.s]
}
job "release" {
  runs_on { runners = "ubuntu-latest" }
  depends_on = [job.build, job.build]
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build, job.release]
}
`,
			errContains: "duplicate needed job",
		},
		{
			name: "job depends_on itself",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  depends_on = [job.build]
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "dependency cycle",
		},
		{
			name: "invalid permissions scope",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  permissions {
    admin = "write"
  }
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "unknown permissions scope",
		},
		{
			name: "invalid permissions level",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  permissions {
    contents = "admin"
  }
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "invalid permission level",
		},
		{
			name: "legacy needs rejected",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  needs = [job.release]
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "unknown attribute 'needs' in job",
		},
		{
			name: "unknown job attribute rejected",
			content: `step "s" { run = "echo hi" }
job "build" {
  runs_on { runners = "ubuntu-latest" }
  myprop = "x"
  steps = [step.s]
}
workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`,
			errContains: "unknown attribute 'myprop' in job",
		},
		{
			name: "unknown action attribute",
			content: `step "echo" {
  run = "echo hi"
}

action "bad" {
  filename = "bad"
  name = "Bad"
  random = "x"

  runs {
    using = "composite"
    steps = [step.echo]
  }
}
`,
			errContains: "unknown attribute 'random' in action",
		},
		{
			name: "unknown action block",
			content: `step "echo" {
  run = "echo hi"
}

action "bad" {
  filename = "bad"
  name = "Bad"

  random {}

  runs {
    using = "composite"
    steps = [step.echo]
  }
}
`,
			errContains: "unknown block 'random' in action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			input := filepath.Join(tmpDir, "in.hcl")

			if err := os.WriteFile(input, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			err := New().Parse(provider.ProviderOps{File: input, OutputDirectory: tmpDir})

			if err == nil {
				t.Fatal("expected parse error but got nil")
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestUnparseValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		errContains string
	}{
		{
			name: "workflow unknown top-level key",
			content: `name: CI
unknown: true
on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`,
			errContains: "unknown key 'unknown' in workflow_yaml",
		},
		{
			name: "workflow has jobs but no on",
			content: `jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`,
			errContains: "workflow_yaml: must define at least one trigger",
		},
		{
			name: "workflow has on but no jobs",
			content: `on:
  push: {}
`,
			errContains: "workflow_yaml: workflow YAML must define both 'on' and 'jobs'",
		},
		{
			name: "job unknown key in yaml",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    mystery: true
    steps:
      - run: echo hi
`,
			errContains: "unknown key 'mystery' in jobs.build",
		},
		{
			name: "step unknown key in yaml",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
        mystery: true
`,
			errContains: "unknown key 'mystery' in jobs.build.steps[0]",
		},
		{
			name: "job has with without uses",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    with:
      target: prod
    steps:
      - run: echo hi
`,
			errContains: "only valid when 'uses'",
		},
		{
			name: "job needs missing reference",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    needs:
      - missing
    steps:
      - run: echo hi
`,
			errContains: "jobs.build.needs: cannot find needed job",
		},
		{
			name: "job steps wrong type",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      run: echo hi
`,
			errContains: "jobs.build: 'steps' must be a list",
		},
		{
			name: "job duplicate needs",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
  release:
    runs-on: ubuntu-latest
    needs: [build, build]
    steps:
      - run: echo hi
`,
			errContains: "duplicate needed job",
		},
		{
			name: "job needs itself",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - run: echo hi
`,
			errContains: "dependency cycle",
		},
		{
			name: "invalid permissions scope in YAML",
			content: `on:
  push: {}
permissions:
  admin: write
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`,
			errContains: "unknown permissions scope",
		},
		{
			name: "invalid permission level in YAML",
			content: `on:
  push: {}
permissions:
  contents: admin
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`,
			errContains: "invalid permission level",
		},
		{
			name: "invalid cron in schedule",
			content: `on:
  schedule:
    - cron: "60 * * * *"
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`,
			errContains: "out of range",
		},
		{
			name: "unclosed expression",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo ${{ broken
`,
			errContains: "unclosed expression",
		},
		{
			name: "invalid uses format",
			content: `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout
`,
			errContains: "must include a version reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			input := filepath.Join(tmpDir, "in.yaml")

			if err := os.WriteFile(input, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			err := New().Unparse(provider.ProviderOps{File: input, OutputDirectory: tmpDir})

			if err == nil {
				t.Fatal("expected unparse error but got nil")
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestUnparseActionValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		errContains string
	}{
		{
			name: "action unknown top-level key",
			content: `name: My Action
description: desc
unknown: true
runs:
  using: composite
  steps:
    - run: echo hi
`,
			errContains: "unknown key 'unknown' in action_yaml",
		},
		{
			name: "action unknown runs key",
			content: `name: My Action
runs:
  using: composite
  unknown: true
  steps:
    - run: echo hi
`,
			errContains: "unknown key 'unknown' in action_yaml.runs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			input := filepath.Join(tmpDir, "action.yaml")

			if err := os.WriteFile(input, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			err := New().Unparse(provider.ProviderOps{File: input, OutputDirectory: tmpDir})

			if err == nil {
				t.Fatal("expected unparse error but got nil")
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}
