// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripHCLFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name: "string literals replaced",
			input: `step "checkout" {
  name = "Checkout"
  run  = "echo hello"
}`,
			contains: []string{`name = "..."`},
			excludes: []string{"Checkout", "echo hello"},
		},
		{
			name: "block labels preserved",
			input: `step "release_app_token" {
  name = "Create release app token"
}`,
			contains: []string{`step "release_app_token"`},
			excludes: []string{"Create release app token"},
		},
		{
			name: "attribute names preserved",
			input: `step "test" {
  name  = "Run tests"
  if    = "always"
  timeout_minutes = "5"
}`,
			contains: []string{"name =", "if =", "timeout_minutes ="},
			excludes: []string{"Run tests", "always"},
		},
		{
			name: "nested blocks preserved",
			input: `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "abc123"
  }
  with {
    name  = "fetch-depth"
    value = "0"
  }
}`,
			contains: []string{"uses {", "with {", `step "checkout"`},
			excludes: []string{"actions/checkout", "abc123", "fetch-depth"},
		},
		{
			name: "comments stripped",
			input: `// This is a secret comment about internal infrastructure
// team@yld.io maintains this
step "test" {
  name = "Test"
}`,
			contains: []string{`step "test"`},
			excludes: []string{"secret comment", "team@yld.io", "infrastructure"},
		},
		{
			name: "heredoc content stripped",
			input: `step "deploy" {
  run = <<EOF
set -euo pipefail
echo "deploying to production.internal.yld.io"
curl -H "Authorization: Bearer $TOKEN" https://api.internal.yld.io/deploy
EOF
}`,
			contains: []string{`run = "..."`},
			excludes: []string{"production.internal", "Bearer", "api.internal"},
		},
		{
			name: "multiple blocks preserved",
			input: `variable "list_os" {
  value = ["ubuntu", "macos"]
}

step "checkout" {
  name = "Checkout"
}

step "test" {
  name = "Test"
}`,
			contains: []string{`variable "list_os"`, `step "checkout"`, `step "test"`},
			excludes: []string{"ubuntu", "macos", "Checkout", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHCLFile([]byte(tt.input), "test.hcl")

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q\ngot:\n%s", want, got)
				}
			}

			for _, excluded := range tt.excludes {
				if strings.Contains(got, excluded) {
					t.Errorf("expected output to NOT contain %q\ngot:\n%s", excluded, got)
				}
			}
		})
	}
}

func TestStripHCLContext(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "steps.hcl"), []byte(`step "checkout" {
  name = "Checkout"
  uses {
    action  = "actions/checkout"
    version = "abc123"
  }
}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "variables.hcl"), []byte(`variable "list_os" {
  value = ["ubuntu", "macos"]
}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Non-HCL file should be ignored
	err = os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Secret docs"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, truncated := StripHCLContext(dir)

	if truncated {
		t.Error("expected no truncation for small input")
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	if !strings.Contains(result, `step "checkout"`) {
		t.Error("expected step block label in result")
	}

	if !strings.Contains(result, `variable "list_os"`) {
		t.Error("expected variable block label in result")
	}

	if strings.Contains(result, "actions/checkout") {
		t.Error("expected action value to be stripped")
	}

	if strings.Contains(result, "ubuntu") {
		t.Error("expected variable value to be stripped")
	}

	if strings.Contains(result, "Secret docs") {
		t.Error("expected non-HCL files to be ignored")
	}
}

func TestStripHCLContextNonexistentDir(t *testing.T) {
	result, truncated := StripHCLContext("/nonexistent/path")

	if result != "" {
		t.Error("expected empty result for nonexistent dir")
	}

	if truncated {
		t.Error("expected no truncation for nonexistent dir")
	}
}

func TestStripHCLContextTruncation(t *testing.T) {
	dir := t.TempDir()

	// Create a large HCL file that exceeds maxContextBytes
	var large strings.Builder
	for i := range maxContextTokens {
		large.WriteString(fmt.Sprintf("step \"step_%d\" {\n  name = \"Step %d\"\n}\n\n", i, i))
	}

	err := os.WriteFile(filepath.Join(dir, "large.hcl"), []byte(large.String()), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, truncated := StripHCLContext(dir)

	if !truncated {
		t.Error("expected truncation for large input")
	}

	if len(result) > maxContextBytes {
		t.Errorf("expected result to be at most %d bytes, got %d", maxContextBytes, len(result))
	}
}

func TestStripHCLFileRealFixture(t *testing.T) {
	// Test against the real steps.hcl fixture if it exists
	fixturePath := "../../cinzel/steps.hcl"
	src, err := os.ReadFile(fixturePath)

	if err != nil {
		t.Skip("cinzel/steps.hcl not found, skipping real fixture test")
	}

	got := stripHCLFile(src, "steps.hcl")

	if got == "" {
		t.Fatal("expected non-empty result from real fixture")
	}

	// Verify no sensitive patterns survive
	sensitivePatterns := []string{
		"secrets.",
		"@yld",
		"yldio",
		"sk-ant-",
		"RELEASE_APP_ID",
		"RELEASE_PRIVATE_KEY",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(got, pattern) {
			t.Errorf("sensitive pattern %q found in stripped output:\n%s", pattern, got)
		}
	}

	// Verify structural elements preserved
	if !strings.Contains(got, "step") {
		t.Error("expected 'step' block type in stripped output")
	}
}
