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

// A nested map is written as an HCL block only where the schema in config.go
// declares one. Guessing from the value's shape produced HCL the parser then
// rejected, and a template body written by the generic writer turned its rule
// blocks into an attribute.
func TestNestedMapsFollowTheHCLSchema(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		wantHCL  []string
		wantYAML []string
	}{
		{
			name: "cache key is an attribute",
			yaml: `build:
  script: [make]
  cache:
    key:
      files: [go.sum]
`,
			wantHCL:  []string{"key = {"},
			wantYAML: []string{"key:", "- go.sum"},
		},
		{
			name: "service variables is an attribute",
			yaml: `build:
  script: [make]
  services:
    - name: postgres:16
      variables:
        POSTGRES_DB: test
`,
			wantHCL:  []string{"variables = {"},
			wantYAML: []string{"POSTGRES_DB: test"},
		},
		{
			name: "default retry is an attribute",
			yaml: `default:
  retry:
    max: 2
build:
  script: [make]
`,
			wantHCL:  []string{"retry = {"},
			wantYAML: []string{"max: 2"},
		},
		{
			name: "include inputs is an attribute",
			yaml: `include:
  - component: gitlab.com/c/t@1
    inputs:
      stage: test
build:
  script: [make]
`,
			wantHCL:  []string{"inputs = {"},
			wantYAML: []string{"stage: test"},
		},
		{
			name: "artifacts reports stays a block",
			yaml: `build:
  script: [make]
  artifacts:
    reports:
      junit: report.xml
`,
			wantHCL:  []string{"reports {"},
			wantYAML: []string{"junit: report.xml"},
		},
		{
			name: "template rules stay blocks",
			yaml: `.base:
  rules:
    - if: always
  cache:
    paths: [.cache]
build:
  extends: [.base]
  script: [make]
`,
			wantHCL:  []string{"rule {", "cache {"},
			wantYAML: []string{"rules:", "- if: always"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			in := filepath.Join(tmp, ".gitlab-ci.yml")
			outDir := filepath.Join(tmp, "hcl")
			backDir := filepath.Join(tmp, "yaml")

			if err := os.WriteFile(in, []byte(tc.yaml), 0o644); err != nil {
				t.Fatal(err)
			}

			p := New()

			if err := p.Unparse(provider.ProviderOps{File: in, OutputDirectory: outDir}); err != nil {
				t.Fatalf("Unparse() error = %v", err)
			}

			hclPath := filepath.Join(outDir, ".gitlab-ci.hcl")
			got, err := os.ReadFile(hclPath)
			if err != nil {
				t.Fatal(err)
			}

			for _, want := range tc.wantHCL {
				if !strings.Contains(string(got), want) {
					t.Errorf("HCL missing %q:\n%s", want, got)
				}
			}

			// The HCL has to be readable by cinzel's own parser, and the
			// values have to come back.
			if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: backDir}); err != nil {
				t.Fatalf("Parse() error = %v\nHCL:\n%s", err, got)
			}

			back, err := os.ReadFile(filepath.Join(backDir, ".gitlab-ci.yml"))
			if err != nil {
				t.Fatal(err)
			}

			for _, want := range tc.wantYAML {
				if !strings.Contains(string(back), want) {
					t.Errorf("reparsed YAML missing %q:\n%s", want, back)
				}
			}
		})
	}
}
