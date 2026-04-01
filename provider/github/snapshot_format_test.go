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
	tests := []struct {
		name       string
		filename   string
		content    string
		goldenFile string
	}{
		{
			name:     "single on block",
			filename: "pull-request.yaml",
			content: `name: Pull Request
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
`,
			goldenFile: "workflow_unparse.golden.hcl",
		},
		{
			name:     "multiple on blocks have blank line between them",
			filename: "multi-on.yaml",
			content: `name: CI
on:
  push:
    branches: [main]
  pull_request: {}
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
`,
			goldenFile: "workflow_unparse_multi_on.golden.hcl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, tt.filename)
			outputDir := filepath.Join(tmpDir, "out")

			if err := os.WriteFile(inputFile, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			if err := New().Unparse(provider.ProviderOps{File: inputFile, OutputDirectory: outputDir}); err != nil {
				t.Fatal(err)
			}

			base := strings.TrimSuffix(tt.filename, filepath.Ext(tt.filename))
			gotBytes, err := os.ReadFile(filepath.Join(outputDir, base+".hcl"))
			if err != nil {
				t.Fatal(err)
			}

			expectedBytes, err := os.ReadFile(filepath.Join("testdata", "fixtures", "formatting", tt.goldenFile))
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
		{
			name:       "job order preserved in parse direction",
			inputFile:  filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_job_order.hcl"),
			outputFile: "workflow-parse-job-order.yaml",
			expected:   filepath.Join("testdata", "fixtures", "formatting", "workflow_parse_job_order.golden.yaml"),
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
