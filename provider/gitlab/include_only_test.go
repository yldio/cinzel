// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A pipeline may be nothing but includes, which is how a project pulls its
// whole configuration in from elsewhere. Such a file used to be skipped on
// unparse, writing nothing and exiting 0.
func TestIncludeOnlyPipelineIsAPipeline(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, ".gitlab-ci.yml")
	outDir := filepath.Join(tmp, "hcl")
	backDir := filepath.Join(tmp, "yaml")
	yml := `include:
  - component: gitlab.com/comp/tmpl@1.0
    inputs:
      stage: build
  - local: other.yml
`

	if err := os.WriteFile(in, []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Unparse(provider.ProviderOps{File: in, OutputDirectory: outDir}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	hclPath := filepath.Join(outDir, ".gitlab-ci.hcl")
	got, err := os.ReadFile(hclPath)

	if err != nil {
		t.Fatalf("Unparse() wrote no HCL: %v", err)
	}

	for _, want := range []string{"include {", "gitlab.com/comp/tmpl@1.0", `local = "other.yml"`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("HCL missing %q:\n%s", want, got)
		}
	}

	if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: backDir}); err != nil {
		t.Fatalf("Parse() error = %v\nHCL:\n%s", err, got)
	}

	back, err := os.ReadFile(filepath.Join(backDir, ".gitlab-ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"include:", "component: gitlab.com/comp/tmpl@1.0", "local: other.yml"} {
		if !strings.Contains(string(back), want) {
			t.Errorf("reparsed YAML missing %q:\n%s", want, back)
		}
	}
}

// A document that is not a pipeline at all is still skipped.
func TestNonPipelineDocumentIsStillSkipped(t *testing.T) {
	tmp := t.TempDir()
	outDir := filepath.Join(tmp, "hcl")

	if err := os.WriteFile(filepath.Join(tmp, "compose.yml"), []byte("version: \"3\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: filepath.Join(tmp, "compose.yml"), OutputDirectory: outDir}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "compose.hcl")); err == nil {
		t.Error("Unparse() wrote HCL for a document that is not a pipeline")
	}
}
