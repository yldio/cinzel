// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

func TestUnparseFormattingSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "pull-request.yaml")
	outputDir := filepath.Join(tmpDir, "out")

	content := `name: Pull Request
on:
  pull_request: {}
jobs:
  pull_request:
    name: ${{ matrix.os }}
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os:
          - ubuntu-24.04
          - macos-15
          - windows-2022
    steps:
      - id: checkout
        name: Checkout
        uses: actions/checkout@v4
`

	if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir}); err != nil {
		t.Fatal(err)
	}

	gotBytes, err := os.ReadFile(filepath.Join(outputDir, "pull-request.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	expectedBytes, err := os.ReadFile(filepath.Join("testdata", "fixtures", "formatting", "workflow_unparse.golden.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	got := normalizeLineEndings(string(gotBytes))
	expected := normalizeLineEndings(string(expectedBytes))

	if got != expected {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- expected ---\n%s", got, expected)
	}
}

func TestParseFormattingSnapshots(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
		expected   string
	}{
		{
			name:       "workflow parse order",
			inputFile:  filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_order.hcl"),
			outputFile: "workflow-parse-order.yaml",
			expected:   filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_order.golden.yaml"),
		},
		{
			name:       "workflow parse expression",
			inputFile:  filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_expression.hcl"),
			outputFile: "workflow-parse-expression.yaml",
			expected:   filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_expression.golden.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := t.TempDir()

			if err := New().Parse(provider.ProviderOps{File: tt.inputFile, OutputDirectory: outDir}); err != nil {
				t.Fatal(err)
			}

			gotBytes, err := os.ReadFile(filepath.Join(outDir, tt.outputFile))
			if err != nil {
				t.Fatal(err)
			}

			expectedBytes, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatal(err)
			}

			got := normalizeLineEndings(string(gotBytes))
			expected := normalizeLineEndings(string(expectedBytes))

			if got != expected {
				t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- expected ---\n%s", got, expected)
			}
		})
	}
}

func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
