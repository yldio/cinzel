// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/command"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/provider"
)

func TestProviderWiringSmoke(t *testing.T) {
	var p provider.Provider = New()

	if p.GetProviderName() != "gitlab" {
		t.Fatalf("provider name = %q, want %q", p.GetProviderName(), "gitlab")
	}

	parseHelpOut := new(bytes.Buffer)
	parseCLI := command.New(parseHelpOut, "v.test")

	if err := parseCLI.Execute([]string{"cinzel", "gitlab", "parse", "--help"}, []provider.Provider{p}); err != nil {
		t.Fatalf("parse help execute error = %v", err)
	}

	if !strings.Contains(parseHelpOut.String(), "Load HCL") {
		t.Fatalf("parse help output missing expected content: %q", parseHelpOut.String())
	}

	unparseHelpOut := new(bytes.Buffer)
	unparseCLI := command.New(unparseHelpOut, "v.test")

	if err := unparseCLI.Execute([]string{"cinzel", "gitlab", "unparse", "--help"}, []provider.Provider{p}); err != nil {
		t.Fatalf("unparse help execute error = %v", err)
	}

	if !strings.Contains(unparseHelpOut.String(), "Load YAML") {
		t.Fatalf("unparse help output missing expected content: %q", unparseHelpOut.String())
	}
}

func TestDefaultOutputDirectories(t *testing.T) {
	opts := provider.ProviderOps{}

	if got := resolveParseOutputDirectory(opts); got != "." {
		t.Fatalf("parse output directory = %q, want %q", got, ".")
	}

	if got := resolveUnparseOutputDirectory(opts); got != "./cinzel" {
		t.Fatalf("unparse output directory = %q, want %q", got, "./cinzel")
	}
}

func TestResolveInputPathValidation(t *testing.T) {

	if _, err := resolveInputPath(provider.ProviderOps{}); err == nil {
		t.Fatal("expected error when file and directory are empty")
	}

	if _, err := resolveInputPath(provider.ProviderOps{File: "a.hcl", Directory: "./in"}); err == nil {
		t.Fatal("expected conflict error when file and directory are both set")
	}
}

func TestDollarBracedEscapePrototype(t *testing.T) {
	expr, diags := hclsyntax.ParseExpression([]byte(`"$${CI_VARIABLE}"`), "", hcl.Pos{})

	if diags.HasErrors() {
		t.Fatalf("ParseExpression() diagnostics = %v", diags)
	}

	hp := hclparser.New(expr, hclparser.NewHCLVars())

	if err := hp.Parse(); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if got := hp.Result().AsString(); got != "${CI_VARIABLE}" {
		t.Fatalf("escaped expression = %q, want %q", got, "${CI_VARIABLE}")
	}
}

func TestParsePipelineBasicFeatures(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "pipeline.hcl")
	outputDir := filepath.Join(tmpDir, "out")

	hclIn := `
stages = ["build", "test"]

variable "deploy_env" {
  name        = "DEPLOY_ENV"
  value       = "production"
  description = "Target environment"
}

job "build" {
  stage  = "build"
  image  = "golang:1.26"
  script = ["go build -o app ./..."]

  artifacts {
    paths     = ["app"]
    expire_in = "1 hour"

    reports {
      junit = ["report.xml"]
    }
  }
}

job "test" {
  stage      = "test"
  depends_on = [job.build]
  script     = ["go test ./..."]
  before_script = ["echo start"]
  after_script  = ["echo done"]
  tags          = ["docker"]

  cache {
    key   = "go-modules"
    paths = ["vendor/"]
  }

	  rule {
	    if   = "$${CI_PIPELINE_SOURCE} == \"merge_request_event\""
	    when = "on_success"
	  }

	  rule {
	    if   = "$CI_PIPELINE_SOURCE == \"push\""
	    when = "never"
	  }
}

workflow {
  rule {
    if   = "$${CI_COMMIT_BRANCH} == \"main\""
    when = "always"
  }
}
`

	if err := os.WriteFile(inputFile, []byte(hclIn), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Parse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	outPath := filepath.Join(outputDir, ".gitlab-ci.yml")
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	out := string(content)
	checks := []string{
		"stages:",
		"variables:",
		"DEPLOY_ENV:",
		"description: Target environment",
		"build:",
		"test:",
		"needs:",
		"- build",
		"rules:",
		"${CI_PIPELINE_SOURCE}",
		"$CI_PIPELINE_SOURCE == \"push\"",
		"artifacts:",
		"reports:",
		"cache:",
		"workflow:",
	}

	for _, check := range checks {

		if !strings.Contains(out, check) {
			t.Fatalf("expected output to contain %q, got:\n%s", check, out)
		}
	}
}

func TestParseDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "pipeline.hcl")

	hclIn := `
stages = ["build"]

job "build" {
  stage  = "build"
  script = ["echo hi"]
}
`

	if err := os.WriteFile(inputFile, []byte(hclIn), 0o644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		err := New().Parse(provider.ProviderOps{File: inputFile, DryRun: true})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
	})

	if !strings.Contains(out, "# file: .gitlab-ci.yml") {
		t.Fatalf("expected dry-run path in stdout, got %q", out)
	}
}

func TestUnparsePipelineBasicFeatures(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, ".gitlab-ci.yml")
	outputDir := filepath.Join(tmpDir, "out")

	yml := `
stages:
  - build
  - test
default:
  image: golang:1.26
include:
  local: .gitlab/base.yml
variables:
  DEPLOY_ENV:
    value: production
    description: Target environment
workflow:
  rules:
    - if: "${CI_COMMIT_BRANCH} == \"main\""
      when: always
build:
  stage: build
  script:
    - go build -o app ./...
test:
  stage: test
  needs:
    - build
  script:
    - go test ./...
  rules:
    - if: "${CI_PIPELINE_SOURCE} == \"merge_request_event\""
      when: on_success
.go_base:
  image: golang:1.26
`

	if err := os.WriteFile(inputFile, []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Unparse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	outPath := filepath.Join(outputDir, ".gitlab-ci.hcl")
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	out := string(content)
	checks := []string{
		"stages = [",
		"variable \"deploy_env\"",
		"DEPLOY_ENV",
		"job \"build\"",
		"job \"test\"",
		"depends_on = [",
		"job.build",
		"workflow {",
		"default {",
		"rule {",
		"$${CI_PIPELINE_SOURCE}",
		"include {",
		"template \"go_base\"",
	}

	for _, check := range checks {

		if !strings.Contains(out, check) {
			t.Fatalf("expected output to contain %q, got:\n%s", check, out)
		}
	}
}

func TestParseRejectsStageReferencesInStagesList(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "pipeline.hcl")

	hclIn := `
stages = [job.build, job.test]

job "build" {
  stage  = "build"
  script = ["echo build"]
}
`

	if err := os.WriteFile(inputFile, []byte(hclIn), 0o644); err != nil {
		t.Fatal(err)
	}

	err := New().Parse(provider.ProviderOps{File: inputFile, OutputDirectory: tmpDir})

	if err == nil {
		t.Fatal("expected parse error for stages job references")
	}
}

func TestUnparseDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "pipeline.yml")

	yml := `
stages:
  - build
build:
  stage: build
  script:
    - echo hi
`

	if err := os.WriteFile(inputFile, []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		err := New().Unparse(provider.ProviderOps{File: inputFile, DryRun: true})
		if err != nil {
			t.Fatalf("Unparse() error = %v", err)
		}
	})

	if !strings.Contains(out, "# file: cinzel/pipeline.hcl") {
		t.Fatalf("expected dry-run output path, got %q", out)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		os.Stdout = originalStdout
		_ = r.Close()
		_ = w.Close()
	})

	os.Stdout = w
	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("stdout close error = %v", err)
	}
	os.Stdout = originalStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout error = %v", err)
	}

	if err := r.Close(); err != nil {
		t.Fatalf("stdout close read pipe error = %v", err)
	}

	return string(out)
}
