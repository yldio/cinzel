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

// A job or a default may declare several caches as a list. Unparse used to
// reject the list form with "cache must be an object".
func TestSeveralCaches(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		wantHCL  []string
		wantYAML []string
	}{
		{
			name: "job with two caches",
			yaml: `build:
  script: [make]
  cache:
    - key: gems
      paths: [vendor/ruby]
    - key: node
      paths: [node_modules]
      policy: pull
`,
			wantHCL:  []string{`key   = "gems"`, `key    = "node"`, `policy = "pull"`},
			wantYAML: []string{"- key: gems", "- key: node", "policy: pull"},
		},
		{
			name: "default with two caches",
			yaml: `default:
  cache:
    - key: a
      paths: [x]
    - key: b
      paths: [y]
build:
  script: [make]
`,
			wantHCL:  []string{"cache {", `key   = "a"`, `key   = "b"`},
			wantYAML: []string{"- key: a", "- key: b"},
		},
		{
			name: "single cache stays an object",
			yaml: `build:
  script: [make]
  cache:
    key: gems
    paths: [vendor/ruby]
`,
			wantHCL:  []string{"cache {", `key   = "gems"`},
			wantYAML: []string{"cache:\n    key: gems"},
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
