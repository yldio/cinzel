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

// A "needs" entry is either a job name or an object carrying that name plus
// options. The object form used to be rejected outright on unparse.
func TestNeedsAsObjects(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		wantHCL  []string
		wantYAML []string
	}{
		{
			name: "job with artifacts",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs:
    - job: build
      artifacts: true
`,
			wantHCL:  []string{"need {", "job       = job.build", "artifacts = true"},
			wantYAML: []string{"job: build", "artifacts: true"},
		},
		{
			name: "cross-project need names no local job",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs:
    - job: other
      project: group/proj
      ref: main
`,
			wantHCL:  []string{"job.other", `project = "group/proj"`},
			wantYAML: []string{"job: other", "project: group/proj"},
		},
		{
			// A hyphenated name is sanitized for the HCL label, and the
			// object form used to keep the sanitized name, so the file did
			// not parse back.
			name: "job with a name that needs sanitizing",
			yaml: `build-app:
  script: [make]
test:
  script: [make test]
  needs:
    - job: build-app
      artifacts: true
`,
			wantHCL:  []string{"job       = job.build_app", `id     = "build-app"`},
			wantYAML: []string{"job: build-app"},
		},
		{
			// The string branch does the same remap and must keep doing it.
			name: "string need to a name that needs sanitizing",
			yaml: `build-app:
  script: [make]
test:
  script: [make test]
  needs: [build-app]
`,
			wantHCL:  []string{"depends_on = [", "job.build_app,"},
			wantYAML: []string{"- build-app"},
		},
		{
			name: "plain string list still an attribute",
			yaml: `build:
  script: [make]
test:
  script: [make test]
  needs: [build]
`,
			wantHCL:  []string{"depends_on = [", "job.build,"},
			wantYAML: []string{"- build"},
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

// Accepting the object form must not let through a need that names nothing.
func TestNeedsObjectWithoutAJobIsRejected(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, ".gitlab-ci.hcl")
	hcl := "job \"build\" {\n  script = [\"make\"]\n}\n\njob \"test\" {\n  script = [\"make test\"]\n\n  need {\n    artifacts = true\n  }\n}\n"

	if err := os.WriteFile(in, []byte(hcl), 0o644); err != nil {
		t.Fatal(err)
	}

	err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: filepath.Join(tmp, "out")})

	if err == nil {
		t.Fatal("Parse() error = nil, want an error")
	}

	if !strings.Contains(err.Error(), "needs must contain") {
		t.Errorf("Parse() error = %q, want it to mention the needs entry", err)
	}
}
