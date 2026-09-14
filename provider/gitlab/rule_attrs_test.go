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

// A rule takes "variables" and "needs" as well. The writer copied every key of
// a rule into the rule block, so HCL carrying either was written out and then
// rejected by cinzel's own parser.
func TestRulesCarryVariablesAndNeeds(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		wantHCL  []string
		wantYAML []string
	}{
		{
			name: "job rule",
			yaml: `build:
  script: [make]
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
      variables:
        TARGET: prod
      needs: []
`,
			wantHCL:  []string{"variables = {", `TARGET = "prod"`, "needs = []"},
			wantYAML: []string{"TARGET: prod", "needs: []"},
		},
		{
			name: "workflow rule",
			yaml: `workflow:
  rules:
    - if: $CI_COMMIT_TAG
      variables:
        DEPLOY: "true"
build:
  script: [make]
`,
			wantHCL:  []string{"variables = {", `DEPLOY = "true"`},
			wantYAML: []string{"DEPLOY: \"true\""},
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
