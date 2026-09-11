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

// Job and template names allow characters an HCL identifier does not, so the
// block label is sanitized. The original name has to survive in an "id"
// attribute, along with every "needs" and "extends" that points at it.
func TestJobNamesThatAreNotIdentifiersSurviveRoundtrip(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, ".gitlab-ci.yml")
	outDir := filepath.Join(tmp, "hcl")
	backDir := filepath.Join(tmp, "yaml")

	yml := `stages:
  - build
  - test
.go-base:
  image: golang:1.26
build-app:
  extends:
    - .go-base
  stage: build
  script:
    - make build
test:unit:
  stage: test
  needs:
    - build-app
  script:
    - make test
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

	for _, want := range []string{`id = "build-app"`, `id = "test:unit"`, `id    = "go-base"`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("HCL missing %s:\n%s", want, got)
		}
	}

	if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: backDir}); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	back, err := os.ReadFile(filepath.Join(backDir, ".gitlab-ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"\n.go-base:", "\nbuild-app:", "\ntest:unit:", "- build-app", "- .go-base"} {
		if !strings.Contains(string(back), want) {
			t.Errorf("reparsed YAML missing %q:\n%s", want, back)
		}
	}

	for _, unwanted := range []string{"build_app", "test_unit", "go_base"} {
		if strings.Contains(string(back), unwanted) {
			t.Errorf("reparsed YAML still holds sanitized name %q:\n%s", unwanted, back)
		}
	}
}
