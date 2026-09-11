// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/yldio/cinzel/provider"
)

// A job key containing a dash cannot be an HCL block label, so unparse used to
// rename it and "needs" with it. The key is now kept as an "id" attribute.
func TestJobKeyWithDashSurvivesRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()

	inputYAML := filepath.Join(tmpDir, "dash.yaml")
	unparseDir := filepath.Join(tmpDir, "unparse")
	parseDir := filepath.Join(tmpDir, "parse")

	content := `name: Dash
on:
  push:
jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - run: make test
  deploy:
    needs: build-and-test
    runs-on: ubuntu-latest
    steps:
      - run: make deploy
`

	if err := os.WriteFile(inputYAML, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Unparse(provider.ProviderOps{File: inputYAML, OutputDirectory: unparseDir}); err != nil {
		t.Fatal(err)
	}

	hclFile := filepath.Join(unparseDir, "dash.hcl")

	hclBytes, err := os.ReadFile(hclFile)
	if err != nil {
		t.Fatal(err)
	}

	if got := string(hclBytes); !strings.Contains(got, `id = "build-and-test"`) {
		t.Errorf("job block does not carry the original key:\n%s", got)
	}

	if err := p.Parse(provider.ProviderOps{File: hclFile, OutputDirectory: parseDir}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(parseDir, "dash.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Jobs map[string]struct {
			Needs []string `yaml:"needs"`
		} `yaml:"jobs"`
	}

	if err := yaml.Unmarshal(got, &doc); err != nil {
		t.Fatal(err)
	}

	if _, ok := doc.Jobs["build-and-test"]; !ok {
		t.Errorf("job key was renamed, got keys %v:\n%s", keysOf(doc.Jobs), got)
	}

	deploy, ok := doc.Jobs["deploy"]

	if !ok {
		t.Fatalf("deploy job is missing:\n%s", got)
	}

	if len(deploy.Needs) != 1 || deploy.Needs[0] != "build-and-test" {
		t.Errorf("needs does not point at the original key, got %v:\n%s", deploy.Needs, got)
	}
}

func keysOf[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))

	for k := range m {
		out = append(out, k)
	}

	return out
}
