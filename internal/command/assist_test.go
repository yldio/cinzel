// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitYAMLDocuments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "single document",
			input: "name: test\non:\n  push:",
			want:  1,
		},
		{
			name:  "two documents",
			input: "name: workflow1\non:\n  push:\n---\nname: workflow2\non:\n  pull_request:",
			want:  2,
		},
		{
			name:  "leading separator ignored",
			input: "---\nname: test",
			want:  1,
		},
		{
			name:  "three documents",
			input: "name: a\n---\nname: b\n---\nname: c",
			want:  3,
		},
		{
			name:  "empty input",
			input: "",
			want:  0,
		},
		{
			name:  "whitespace only",
			input: "   \n\n  ",
			want:  0,
		},
		{
			name:  "separator with whitespace",
			input: "name: a\n  ---  \nname: b",
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitYAMLDocuments(tt.input)
			if len(got) != tt.want {
				t.Errorf("splitYAMLDocuments() returned %d documents, want %d\ndocs: %v", len(got), tt.want, got)
			}
		})
	}
}

func TestSplitYAMLDocumentsContent(t *testing.T) {
	input := "name: workflow1\non:\n  push:\n---\nname: workflow2\non:\n  pull_request:"
	docs := splitYAMLDocuments(input)

	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}

	if got := docs[0]; got != "name: workflow1\non:\n  push:\n" {
		t.Errorf("doc[0]:\ngot:  %q\nwant: %q", got, "name: workflow1\non:\n  push:\n")
	}

	if got := docs[1]; got != "name: workflow2\non:\n  pull_request:\n" {
		t.Errorf("doc[1]:\ngot:  %q\nwant: %q", got, "name: workflow2\non:\n  pull_request:\n")
	}
}

func TestSplitHCLBlocksAST(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantBlocks int
	}{
		{
			name: "two blocks",
			input: `step "checkout" {
  name = "Checkout"
}

step "test" {
  name = "Test"
}`,
			wantBlocks: 2,
		},
		{
			name: "block with braces in string",
			input: `step "deploy" {
  run = "echo ${VAR}"
}`,
			wantBlocks: 1,
		},
		{
			name: "nested blocks",
			input: `workflow "pr" {
  on "pull_request" {}
  jobs = [job.test]
}`,
			wantBlocks: 1,
		},
		{
			name:       "invalid HCL falls back to single block",
			input:      "this is not valid HCL {{{",
			wantBlocks: 1,
		},
		{
			name: "top-level attribute",
			input: `variable "os" {
  value = "ubuntu"
}

name = "test"`,
			wantBlocks: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitHCLBlocksAST([]byte(tt.input), "test.hcl")
			if len(got) != tt.wantBlocks {
				t.Errorf("splitHCLBlocksAST() returned %d blocks, want %d\nblocks: %v", len(got), tt.wantBlocks, got)
			}
		})
	}
}

func TestMergeHCLFiles(t *testing.T) {
	dir := t.TempDir()

	file1 := `step "checkout" {
  name = "Checkout"
}

step "test" {
  name = "Test"
}
`
	file2 := `step "checkout" {
  name = "Checkout"
}

step "build" {
  name = "Build"
}
`

	if err := os.WriteFile(filepath.Join(dir, "a.hcl"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "b.hcl"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}

	merged, err := mergeHCLFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	// checkout should appear only once (deduped)
	if count := strings.Count(merged, `step "checkout"`); count != 1 {
		t.Errorf("expected 1 checkout block, got %d\nmerged:\n%s", count, merged)
	}

	// test and build should each appear once
	if !strings.Contains(merged, `step "test"`) {
		t.Error("expected test block in merged output")
	}

	if !strings.Contains(merged, `step "build"`) {
		t.Error("expected build block in merged output")
	}
}

func TestMergeHCLFilesIgnoresNonHCL(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Secret"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "test.hcl"), []byte(`step "a" {}`), 0644); err != nil {
		t.Fatal(err)
	}

	merged, err := mergeHCLFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(merged, "Secret") {
		t.Error("non-HCL content should not appear in merged output")
	}
}

func TestConfirmCost(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "yes", input: "y\n", wantErr: false},
		{name: "YES", input: "YES\n", wantErr: false},
		{name: "no", input: "n\n", wantErr: true},
		{name: "empty", input: "\n", wantErr: true},
		{name: "eof", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := confirmCost(&buf, strings.NewReader(tt.input), "anthropic", "default")

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestResolveAIProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
		wantName string
	}{
		{name: "anthropic explicit", provider: "anthropic", wantErr: true},
		{name: "openai explicit", provider: "openai", wantErr: true},
		{name: "empty defaults to anthropic", provider: "", wantErr: true},
		{name: "unknown", provider: "gemini", wantErr: true},
	}

	// All cases error because no API keys are set in test env.
	// We verify provider resolution logic, not API connectivity.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveAIProvider(tt.provider, "")
			if tt.provider == "gemini" {
				if err == nil || !strings.Contains(err.Error(), "unknown AI provider") {
					t.Errorf("expected unknown provider error, got %v", err)
				}
			} else if err == nil {
				t.Error("expected missing API key error without env var")
			}
		})
	}
}

func TestResolveAIProviderWithKey(t *testing.T) {
	p, err := resolveAIProvider("anthropic", "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Name() != "anthropic" {
		t.Errorf("expected anthropic, got %s", p.Name())
	}

	p, err = resolveAIProvider("openai", "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}
