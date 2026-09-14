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

// roundtripYAML unparses a pipeline and parses the result back, returning the
// emitted HCL and the YAML it produced.
func roundtripYAML(t *testing.T, yml string) (string, string) {
	t.Helper()

	tmp := t.TempDir()
	in := filepath.Join(tmp, ".gitlab-ci.yml")
	outDir := filepath.Join(tmp, "hcl")
	backDir := filepath.Join(tmp, "yaml")

	if err := os.WriteFile(in, []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Unparse(provider.ProviderOps{File: in, OutputDirectory: outDir}); err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	hclPath := filepath.Join(outDir, ".gitlab-ci.hcl")
	hcl, err := os.ReadFile(hclPath)

	if err != nil {
		t.Fatalf("Unparse() wrote no HCL: %v", err)
	}

	if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: backDir}); err != nil {
		t.Fatalf("Parse() error = %v\nHCL:\n%s", err, hcl)
	}

	back, err := os.ReadFile(filepath.Join(backDir, ".gitlab-ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	return string(hcl), string(back)
}

// "only" and "except" are the older way to say "rules". They were written out
// but absent from the schema, so the emitted HCL did not parse back.
func TestLegacyOnlyAndExceptSurvive(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yml      string
		wantYAML []string
	}{
		{
			name: "object form",
			yml: `legacy:
  script: [make]
  only:
    refs: [main, tags]
    variables: ['$CI == "true"']
    changes: [src/**/*]
  except:
    refs: [schedules]
`,
			wantYAML: []string{"only:", "refs:", "- main", "except:", "- schedules"},
		},
		{
			name: "list form",
			yml: `legacy:
  script: [make]
  only: [main]
  except: [schedules]
`,
			wantYAML: []string{"only:", "- main", "except:", "- schedules"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, back := roundtripYAML(t, tc.yml)

			for _, want := range tc.wantYAML {
				if !strings.Contains(back, want) {
					t.Errorf("reparsed YAML missing %q:\n%s", want, back)
				}
			}
		})
	}
}

// A variable's "expand" was dropped on unparse, since the variable writer names
// the keys it copies rather than writing the object generically.
func TestVariableExpandSurvives(t *testing.T) {
	_, back := roundtripYAML(t, `variables:
  DEPLOY:
    value: "false"
    expand: false
    description: whether to deploy
build:
  script: [make]
`)

	for _, want := range []string{"expand: false", "description: whether to deploy", `value: "false"`} {
		if !strings.Contains(back, want) {
			t.Errorf("reparsed YAML missing %q:\n%s", want, back)
		}
	}
}

// A variable that carries nothing but a value stays a plain scalar, rather than
// becoming an object with a lone "value".
func TestPlainVariableStaysAScalar(t *testing.T) {
	_, back := roundtripYAML(t, `variables:
  TAG: latest
build:
  script: [make]
`)

	if !strings.Contains(back, "TAG: latest") {
		t.Errorf("reparsed YAML missing %q:\n%s", "TAG: latest", back)
	}
}

// GitLab takes a single command as a bare string as well as a list, but the
// validator demanded a list, so such a job would not parse back.
func TestStringScriptSurvives(t *testing.T) {
	_, back := roundtripYAML(t, `build:
  script: make
  before_script: setup
  after_script: teardown
`)

	for _, want := range []string{"script: make", "before_script: setup", "after_script: teardown"} {
		if !strings.Contains(back, want) {
			t.Errorf("reparsed YAML missing %q:\n%s", want, back)
		}
	}
}

// An empty script is still rejected, whichever shape it takes. The list case
// lives in trigger_job_test.go; this covers the string the validator now takes.
func TestEmptyStringScriptIsRejected(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, ".gitlab-ci.hcl")

	if err := os.WriteFile(in, []byte("job \"build\" {\n  script = \"\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: filepath.Join(tmp, "out")}); err == nil {
		t.Error("Parse() accepted an empty script")
	}
}
