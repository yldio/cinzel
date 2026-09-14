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

// The writer copies a job's keywords through, so any keyword the HCL schema
// does not declare produces HCL that cinzel's own parser rejects. Every
// keyword GitLab documents has to survive the roundtrip.
func TestKeywordsSurviveTheRoundtrip(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		wantYAML []string
	}{
		{
			name: "job keywords",
			yaml: `build:
  script: [make]
  artifacts:
    paths: [bin/]
test:
  script: [make test]
  dependencies: [build]
  identity: google_cloud
  when: manual
  manual_confirmation: "Are you sure?"
  inherit:
    default: false
    variables: [A, B]
  secrets:
    DB_PASS:
      vault: production/db/password@ops
  id_tokens:
    VAULT_TOKEN:
      aud: https://vault.example.com
  hooks:
    pre_get_sources_script: [echo hi]
  dast_configuration:
    site_profile: Site
    scanner_profile: Scanner
`,
			wantYAML: []string{
				"dependencies:", "identity: google_cloud", "manual_confirmation:",
				"inherit:", "secrets:", "id_tokens:", "hooks:", "dast_configuration:",
			},
		},
		{
			name: "delayed rule",
			yaml: `build:
  script: [make]
  rules:
    - if: $CI
      when: delayed
      start_in: 30 minutes
      interruptible: true
`,
			wantYAML: []string{"start_in: 30 minutes", "interruptible: true"},
		},
		{
			name: "artifacts and cache keys",
			yaml: `build:
  script: [make]
  artifacts:
    paths: [bin/]
    expose_as: binaries
    public: false
    access: developer
  cache:
    key: gems
    paths: [vendor]
    unprotect: true
`,
			wantYAML: []string{"expose_as: binaries", "public: false", "access: developer", "unprotect: true"},
		},
		{
			name: "service runtime keys",
			yaml: `build:
  script: [make]
  services:
    - name: postgres:14
      docker:
        platform: linux/amd64
        user: dev
`,
			wantYAML: []string{"docker:", "platform: linux/amd64"},
		},
		{
			name: "include rules",
			yaml: `include:
  - remote: https://example.com/x.yml
    rules:
      - if: $CI
build:
  script: [make]
`,
			wantYAML: []string{"rules:", "if: $CI"},
		},
		{
			name: "default artifacts, hooks and id_tokens",
			yaml: `default:
  artifacts:
    expire_in: 1 week
  hooks:
    pre_get_sources_script: [echo x]
  id_tokens:
    T:
      aud: a
build:
  script: [make]
`,
			wantYAML: []string{"artifacts:", "expire_in: 1 week", "hooks:", "id_tokens:"},
		},
		{
			name: "variable options and workflow auto_cancel",
			yaml: `variables:
  DEPLOY_ENV:
    value: staging
    description: Target environment
    options: [staging, prod]
workflow:
  name: Pipeline
  auto_cancel:
    on_new_commit: interruptible
build:
  script: [make]
`,
			wantYAML: []string{"options:", "- staging", "auto_cancel:", "on_new_commit: interruptible"},
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

// A job's "inputs" and "publish" are in GitLab's own schema but were missing
// from cinzel's, so the emitted HCL did not parse back.
func TestJobInputsAndPublish(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, ".gitlab-ci.yml")
	outDir := filepath.Join(tmp, "hcl")
	backDir := filepath.Join(tmp, "yaml")
	yml := `build:
  script: [make]
  inputs:
    stage: test
  publish: public
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
		t.Fatal(err)
	}

	if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: backDir}); err != nil {
		t.Fatalf("Parse() error = %v\nHCL:\n%s", err, got)
	}

	back, err := os.ReadFile(filepath.Join(backDir, ".gitlab-ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"stage: test", "publish: public"} {
		if !strings.Contains(string(back), want) {
			t.Errorf("reparsed YAML missing %q:\n%s", want, back)
		}
	}
}
