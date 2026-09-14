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

// A job does not have to carry its own "script": a trigger job has none by
// definition, and an extending job inherits one. Such a job used to be demoted
// to a bare top-level block that cinzel's own parser then rejected.
func TestJobsWithoutTheirOwnScript(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		wantHCL  []string
		wantYAML []string
	}{
		{
			name: "trigger job",
			yaml: `build:
  script: [make]
deploy:
  trigger:
    project: g/p
    branch: main
`,
			wantHCL:  []string{`job "deploy"`, "trigger = {"},
			wantYAML: []string{"\ndeploy:", "project: g/p"},
		},
		{
			name: "job inheriting script through extends",
			yaml: `stages: [build]
.base:
  script: [make]
build:
  extends: [.base]
  stage: build
`,
			wantHCL:  []string{`job "build"`, "template.base"},
			wantYAML: []string{"\nbuild:", "- .base"},
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

// Dropping the script requirement must not let through a job that has no way
// to get one.
func TestJobWithNoScriptAndNothingToInheritIsRejected(t *testing.T) {
	for _, tc := range []struct{ name, hcl, want string }{
		{"no script at all", "job \"build\" {\n  stage = \"build\"\n}\n", "must define 'script'"},
		{"empty script", "job \"build\" {\n  script = []\n}\n", "non-empty string or list"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			in := filepath.Join(tmp, ".gitlab-ci.hcl")

			if err := os.WriteFile(in, []byte(tc.hcl), 0o644); err != nil {
				t.Fatal(err)
			}

			err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: filepath.Join(tmp, "out")})

			if err == nil {
				t.Fatal("Parse() error = nil, want an error")
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("Parse() error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}
