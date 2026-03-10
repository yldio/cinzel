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

func TestGitHubMetadata(t *testing.T) {
	p := New()

	if p.GetProviderName() != "github" {
		t.Fatalf("unexpected provider name: %s", p.GetProviderName())
	}

	if p.GetDescription() == "" {
		t.Fatal("provider description should not be empty")
	}

	if p.GetParseDescription() == "" {
		t.Fatal("parse description should not be empty")
	}

	if p.GetUnparseDescription() == "" {
		t.Fatal("unparse description should not be empty")
	}
}

func TestParseAndUnparse(t *testing.T) {
	t.Run("parses workflow with jobs and referenced steps", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "workflow.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "checkout" {
  uses {
    action = "actions/checkout"
    version = "v4"
  }
}

step "test" {
  run = "go test ./..."
}

job "build" {
  name = "Build"
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.checkout,
    step.test,
  ]
}

workflow "ci" {
  filename = "ci"

  on "push" {
    branches = ["main"]
  }

  jobs = [
    job.build,
  ]
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		p := New()
		err := p.Parse(provider.ProviderOps{
			File:            inputFile,
			OutputDirectory: outputDir,
		})
		if err != nil {
			t.Fatal(err)
		}

		outputFile := filepath.Join(outputDir, "ci.yaml")
		b, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		out := string(b)

		if !strings.Contains(out, "jobs:") {
			t.Fatalf("expected jobs section in workflow output, got: %q", out)
		}

		if !strings.Contains(out, "uses: actions/checkout@v4") {
			t.Fatalf("expected referenced step rendered in workflow output, got: %q", out)
		}

		if !strings.Contains(out, "runs-on: ubuntu-latest") {
			t.Fatalf("expected runs-on in workflow output, got: %q", out)
		}
	})

	t.Run("keeps workflow top-level key order", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "ordered.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }
  steps = [step.echo]
}

workflow "release" {
  filename = "ordered"
  name = "Build Release"

  on "release" {
    types = ["created"]
  }

  jobs = [job.build]
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Parse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		outBytes, err := os.ReadFile(filepath.Join(outputDir, "ordered.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		out := string(outBytes)
		nameIdx := strings.Index(out, "name: Build Release")
		onIdx := strings.Index(out, "on:")
		jobsIdx := strings.Index(out, "jobs:")

		if nameIdx == -1 || onIdx == -1 || jobsIdx == -1 {
			t.Fatalf("expected name, on and jobs keys in output, got:\n%s", out)
		}

		if !(nameIdx < onIdx && onIdx < jobsIdx) {
			t.Fatalf("expected name -> on -> jobs order, got:\n%s", out)
		}

		if strings.Contains(out, `"on":`) {
			t.Fatalf("expected unquoted on key, got:\n%s", out)
		}
	})

	t.Run("renders empty on events without inline empty object", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "pull-request.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }
  steps = [step.echo]
}

workflow "pr" {
  filename = "pull-request"
  name = "Pull Request"
  on "pull_request" {}
  jobs = [job.build]
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Parse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		outBytes, err := os.ReadFile(filepath.Join(outputDir, "pull-request.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		out := string(outBytes)

		if !strings.Contains(out, "pull_request:\n") {
			t.Fatalf("expected pull_request key without inline object, got:\n%s", out)
		}

		if strings.Contains(out, "pull_request: {}") {
			t.Fatalf("expected no inline empty object for pull_request, got:\n%s", out)
		}
	})

	t.Run("parses workflow_call output blocks with labels", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "workflow-call.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "build" {
  run = "echo build"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [
    step.build,
  ]
}

workflow "ci" {
  filename = "ci"

  on "workflow_call" {
    output "artifact-url" {
      description = "Artifact URL"
      value = "$${{ jobs.build.outputs.artifact }}"
    }
  }

  jobs = [
    job.build,
  ]
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		p := New()
		err := p.Parse(provider.ProviderOps{
			File:            inputFile,
			OutputDirectory: outputDir,
		})
		if err != nil {
			t.Fatal(err)
		}

		outputFile := filepath.Join(outputDir, "ci.yaml")
		b, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		out := string(b)

		if !strings.Contains(out, "workflow_call:") {
			t.Fatalf("expected workflow_call trigger in output, got: %q", out)
		}

		if !strings.Contains(out, "artifact-url:") {
			t.Fatalf("expected labeled output under workflow_call, got: %q", out)
		}
	})

	t.Run("parses strategy matrix variable list into GitHub matrix keys", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "matrix.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  strategy {
    matrix {
      variable = [
        {
          name = "goos"
          value = ["linux", "windows", "darwin"]
        },
        {
          name = "goarch"
          value = ["386", "amd64", "arm64"]
        },
      ]

      exclude = [
        {
          goos = "darwin"
          goarch = "386"
        },
      ]
    }
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "matrix"
  on "push" {}
  jobs = [job.build]
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Parse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		outBytes, err := os.ReadFile(filepath.Join(outputDir, "matrix.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		out := string(outBytes)

		if !strings.Contains(out, "goos:") || !strings.Contains(out, "goarch:") {
			t.Fatalf("expected matrix keys goos/goarch, got:\n%s", out)
		}

		if strings.Contains(out, "variable:") {
			t.Fatalf("expected matrix variable field to be normalized away, got:\n%s", out)
		}
	})

	t.Run("parses strategy matrix single variable block", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "matrix-single.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "echo" {
  run = "echo hi"
}

job "pull_request" {
  runs_on {
    runners = "ubuntu-latest"
  }

  strategy {
    matrix {
      variable {
        name = "goos"
        value = ["linux", "darwin"]
      }
    }
  }

  steps = [step.echo]
}

workflow "pr" {
  filename = "matrix-single"
  on "pull_request" {}
  jobs = [job.pull_request]
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Parse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		outBytes, err := os.ReadFile(filepath.Join(outputDir, "matrix-single.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		out := string(outBytes)

		if !strings.Contains(out, "goos:") {
			t.Fatalf("expected single matrix variable to be normalized into key, got:\n%s", out)
		}

		if strings.Contains(out, "variable:") {
			t.Fatalf("expected matrix variable key to be removed, got:\n%s", out)
		}
	})

	t.Run("parses step blocks into yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "steps.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "build" {
  name = "Build"
  run = "go test ./..."
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		p := New()
		err := p.Parse(provider.ProviderOps{
			File:            inputFile,
			OutputDirectory: outputDir,
		})
		if err != nil {
			t.Fatal(err)
		}

		outputFile := filepath.Join(outputDir, "steps.yaml")
		b, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		out := string(b)

		if !strings.Contains(out, "build:") {
			t.Fatalf("expected step identifier in yaml output, got: %q", out)
		}

		if !strings.Contains(out, "run: go test ./...") {
			t.Fatalf("expected run command in yaml output, got: %q", out)
		}
	})

	t.Run("unparses yaml step map into hcl", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "steps.yaml")
		outputDir := filepath.Join(tmpDir, "out")

		content := `build:
  name: Build
  run: go test ./...
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		p := New()
		err := p.Unparse(provider.ProviderOps{
			File:            inputFile,
			OutputDirectory: outputDir,
		})
		if err != nil {
			t.Fatal(err)
		}

		outputFile := filepath.Join(outputDir, "steps.hcl")
		b, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatal(err)
		}

		out := string(b)
		if !strings.Contains(out, `step "build"`) {
			t.Fatalf("expected step block in hcl output, got: %q", out)
		}

		if !strings.Contains(out, `run = "go test ./..."`) {
			t.Fatalf("expected run attribute in hcl output, got: %q", out)
		}
	})

	t.Run("unparses workflow yaml into workflow, jobs and steps", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "ci.yaml")
		outputDir := filepath.Join(tmpDir, "out")

		content := `name: CI
on:
  push:
    branches:
      - main
jobs:
  build-test:
    runs-on: ubuntu-latest
    steps:
      - id: checkout
        uses: actions/checkout@v4
      - name: Test
        run: go test ./...
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		p := New()
		err := p.Unparse(provider.ProviderOps{
			File:            inputFile,
			OutputDirectory: outputDir,
		})
		if err != nil {
			t.Fatal(err)
		}

		hclOutPath := filepath.Join(outputDir, "ci.hcl")
		hclBytes, err := os.ReadFile(hclOutPath)
		if err != nil {
			t.Fatal(err)
		}

		hclOut := string(hclBytes)
		if !strings.Contains(hclOut, `workflow "ci"`) {
			t.Fatalf("expected workflow block in unparsed output, got: %q", hclOut)
		}

		if !strings.Contains(hclOut, `job "build_test"`) {
			t.Fatalf("expected sanitized job block in unparsed output, got: %q", hclOut)
		}

		if !strings.Contains(hclOut, `step "checkout"`) {
			t.Fatalf("expected generated step block in unparsed output, got: %q", hclOut)
		}

		if !strings.Contains(hclOut, `steps = [`) {
			t.Fatalf("expected job step references in unparsed output, got: %q", hclOut)
		}

		if !strings.Contains(hclOut, "step \"checkout\" {\n  id = \"checkout\"\n\n  uses") {
			t.Fatalf("expected blank line between step properties, got: %q", hclOut)
		}

		parseOutDir := filepath.Join(tmpDir, "parsed")
		err = p.Parse(provider.ProviderOps{
			File:            hclOutPath,
			OutputDirectory: parseOutDir,
		})
		if err != nil {
			t.Fatal(err)
		}

		yamlOutPath := filepath.Join(parseOutDir, "ci.yaml")
		yamlBytes, err := os.ReadFile(yamlOutPath)
		if err != nil {
			t.Fatal(err)
		}

		yamlOut := string(yamlBytes)

		if !strings.Contains(yamlOut, "jobs:") {
			t.Fatalf("expected jobs section after roundtrip parse, got: %q", yamlOut)
		}

		if !strings.Contains(yamlOut, "uses: actions/checkout@v4") {
			t.Fatalf("expected step uses after roundtrip parse, got: %q", yamlOut)
		}
	})

	t.Run("unparses matrix axes into variable blocks with references", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "matrix.yaml")
		outputDir := filepath.Join(tmpDir, "out")

		content := `name: Pull Request
on:
  pull_request:
jobs:
  pull_request:
    name: ${{ matrix.os }}
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os:
          - ubuntu-24.04
          - macos-15
          - windows-2022
    steps:
      - id: checkout
        name: Checkout
        uses: actions/checkout@v4
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Unparse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		hclBytes, err := os.ReadFile(filepath.Join(outputDir, "matrix.hcl"))
		if err != nil {
			t.Fatal(err)
		}

		hclOut := string(hclBytes)

		if !strings.Contains(hclOut, "variable {") {
			t.Fatalf("expected matrix variable block, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, `name  = "os"`) {
			t.Fatalf("expected matrix variable name, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, "value = variable.list_os") {
			t.Fatalf("expected matrix variable reference value, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, `variable "list_os"`) {
			t.Fatalf("expected generated root variable list_os, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, "runs_on {\n    runners = \"$${{ matrix.os }}\"\n  }\n\n  strategy {") {
			t.Fatalf("expected single blank line between runs_on and strategy, got:\n%s", hclOut)
		}

	})

	t.Run("formats matrix sections with blank line between exclude and variable", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "matrix-sections.yaml")
		outputDir := filepath.Join(tmpDir, "out")

		content := `name: Release Go Binary
on:
  push: {}
jobs:
  releases_matrix:
    name: Release Go Binary
    runs-on: ubuntu-latest
    strategy:
      matrix:
        exclude:
          - goos: darwin
            goarch: "386"
          - goos: windows
            goarch: arm64
        goos:
          - linux
          - windows
          - darwin
        goarch:
          - "386"
          - amd64
          - arm64
    steps:
      - id: checkout
        uses: actions/checkout@v4
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Unparse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		hclBytes, err := os.ReadFile(filepath.Join(outputDir, "matrix-sections.hcl"))
		if err != nil {
			t.Fatal(err)
		}

		hclOut := string(hclBytes)

		if !strings.Contains(hclOut, "runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  strategy {") {
			t.Fatalf("expected single blank line between runs_on and strategy, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, "exclude = [") || !strings.Contains(hclOut, "}]\n\n      variable {") {
			t.Fatalf("expected blank line between matrix exclude and variable sections, got:\n%s", hclOut)
		}
	})

	t.Run("formats workflow block with spacing inline empty on and trailing commas", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "pull-request.yaml")
		outputDir := filepath.Join(tmpDir, "out")

		content := `name: Pull Request
on:
  pull_request: {}
jobs:
  pull_request:
    name: ${{ matrix.os }}
    uses: org/repo/.github/workflows/reusable.yaml@v1
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Unparse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		hclBytes, err := os.ReadFile(filepath.Join(outputDir, "pull-request.hcl"))
		if err != nil {
			t.Fatal(err)
		}

		hclOut := string(hclBytes)

		if !strings.Contains(hclOut, "filename = \"pull-request\"\n\n  name") {
			t.Fatalf("expected blank line between filename and name, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, "on \"pull_request\" {") {
			t.Fatalf("expected pull_request on block, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, "jobs = [\n    job.pull_request,\n  ]") {
			t.Fatalf("expected trailing comma in jobs list, got:\n%s", hclOut)
		}

		if !strings.Contains(hclOut, "job \"pull_request\" {\n  name = \"$${{ matrix.os }}\"\n\n  uses") {
			t.Fatalf("expected blank line between job properties, got:\n%s", hclOut)
		}
	})

	t.Run("maps depends_on in HCL to needs in YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "depends-on.hcl")
		outputDir := filepath.Join(tmpDir, "out")

		content := `step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

job "release" {
  runs_on {
    runners = "ubuntu-latest"
  }

  depends_on = [job.build]
  steps = [step.echo]
}

workflow "ci" {
  filename = "depends-on"
  on "push" {}
  jobs = [job.build, job.release]
}
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Parse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		outBytes, err := os.ReadFile(filepath.Join(outputDir, "depends-on.yaml"))
		if err != nil {
			t.Fatal(err)
		}

		out := string(outBytes)

		if !strings.Contains(out, "needs:\n      - build") {
			t.Fatalf("expected YAML needs mapped from depends_on, got:\n%s", out)
		}
	})

	t.Run("maps YAML needs to depends_on in HCL", func(t *testing.T) {
		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "needs.yaml")
		outputDir := filepath.Join(tmpDir, "out")

		content := `on:
  push: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo build
  release:
    runs-on: ubuntu-latest
    needs:
      - build
    steps:
      - run: echo release
`

		if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := New().Unparse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir})
		if err != nil {
			t.Fatal(err)
		}

		outBytes, err := os.ReadFile(filepath.Join(outputDir, "needs.hcl"))
		if err != nil {
			t.Fatal(err)
		}

		out := string(outBytes)

		if !strings.Contains(out, "depends_on = [") {
			t.Fatalf("expected depends_on in HCL output, got:\n%s", out)
		}

		if strings.Contains(out, "needs = [") {
			t.Fatalf("did not expect needs in HCL output, got:\n%s", out)
		}
	})
}

func TestResolveInputPath(t *testing.T) {
	_, err := resolveInputPath(provider.ProviderOps{})

	if err == nil {
		t.Fatal("expected validation error when no input is set")
	}

	_, err = resolveInputPath(provider.ProviderOps{File: "a", Directory: "b"})

	if err == nil {
		t.Fatal("expected validation error when both file and directory are set")
	}
}
