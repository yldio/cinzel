// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/yldio/cinzel/provider"
)

func TestParseFixtureMatrixValid(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "fixtures", "matrix", "parse", "valid", "*.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".hcl")
		t.Run(name, func(t *testing.T) {
			outDir := t.TempDir()

			if err := New().Parse(provider.ProviderOps{File: input, OutputDirectory: outDir}); err != nil {
				t.Fatal(err)
			}

			got, err := os.ReadFile(filepath.Join(outDir, ".gitlab-ci.yml"))
			if err != nil {
				t.Fatal(err)
			}

			expected, err := os.ReadFile(strings.TrimSuffix(input, ".hcl") + ".golden.yaml")
			if err != nil {
				t.Fatal(err)
			}

			assertYAMLSemanticEqual(t, got, expected)
		})
	}
}

func TestParseFixtureMatrixInvalid(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "fixtures", "matrix", "parse", "invalid", "*.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".hcl")
		t.Run(name, func(t *testing.T) {
			expectedErrBytes, err := os.ReadFile(strings.TrimSuffix(input, ".hcl") + ".error.txt")
			if err != nil {
				t.Fatal(err)
			}

			parseErr := New().Parse(provider.ProviderOps{File: input, OutputDirectory: t.TempDir()})

			if parseErr == nil {
				t.Fatal("expected parse error")
			}

			expectedErr := strings.TrimSpace(string(expectedErrBytes))

			if !strings.Contains(parseErr.Error(), expectedErr) {
				t.Fatalf("expected error containing %q, got %q", expectedErr, parseErr.Error())
			}
		})
	}
}

func TestUnparseFixtureMatrixValid(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "fixtures", "matrix", "unparse", "valid", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range inputs {
		if strings.HasSuffix(input, ".roundtrip.golden.yaml") {
			continue
		}

		name := strings.TrimSuffix(filepath.Base(input), ".yaml")
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			unparseDir := filepath.Join(tmpDir, "unparse")
			parseDir := filepath.Join(tmpDir, "parse")

			if err := New().Unparse(provider.ProviderOps{File: input, OutputDirectory: unparseDir}); err != nil {
				t.Fatal(err)
			}

			hclPath := filepath.Join(unparseDir, name+".hcl")

			if err := New().Parse(provider.ProviderOps{File: hclPath, OutputDirectory: parseDir}); err != nil {
				t.Fatal(err)
			}

			got, err := os.ReadFile(filepath.Join(parseDir, ".gitlab-ci.yml"))
			if err != nil {
				t.Fatal(err)
			}

			expected, err := os.ReadFile(strings.TrimSuffix(input, ".yaml") + ".roundtrip.golden.yaml")
			if err != nil {
				t.Fatal(err)
			}

			assertYAMLSemanticEqual(t, got, expected)
		})
	}
}

func TestUnparseFixtureMatrixInvalid(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "fixtures", "matrix", "unparse", "invalid", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".yaml")
		t.Run(name, func(t *testing.T) {
			expectedErrBytes, err := os.ReadFile(strings.TrimSuffix(input, ".yaml") + ".error.txt")
			if err != nil {
				t.Fatal(err)
			}

			unparseErr := New().Unparse(provider.ProviderOps{File: input, OutputDirectory: t.TempDir()})

			if unparseErr == nil {
				t.Fatal("expected unparse error")
			}

			expectedErr := strings.TrimSpace(string(expectedErrBytes))

			if !strings.Contains(unparseErr.Error(), expectedErr) {
				t.Fatalf("expected error containing %q, got %q", expectedErr, unparseErr.Error())
			}
		})
	}
}

func assertYAMLSemanticEqual(t *testing.T, got []byte, expected []byte) {
	t.Helper()

	var gotDoc any

	if err := yaml.Unmarshal(got, &gotDoc); err != nil {
		t.Fatalf("failed to unmarshal actual YAML: %v\n---\n%s", err, string(got))
	}

	var expectedDoc any

	if err := yaml.Unmarshal(expected, &expectedDoc); err != nil {
		t.Fatalf("failed to unmarshal expected YAML: %v\n---\n%s", err, string(expected))
	}

	if !reflect.DeepEqual(gotDoc, expectedDoc) {
		t.Fatalf("YAML documents differ semantically\n--- got ---\n%s\n--- expected ---\n%s", string(got), string(expected))
	}
}
