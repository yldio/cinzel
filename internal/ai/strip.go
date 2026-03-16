// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	maxContextTokens  = 8000
	bytesPerToken     = 4
	maxContextBytes   = maxContextTokens * bytesPerToken
	strippedLiteral   = `"..."`
	strippedHeredoc   = "..."
)

// StripHCLContext reads HCL files from the given directory, strips all string
// literal values, heredoc content, and comments, then returns the structural
// skeleton. Returns an empty string if the directory doesn't exist or has no
// HCL files.
func StripHCLContext(dir string) (string, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}

	var parts []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hcl") {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		stripped := stripHCLFile(content, entry.Name())
		if stripped != "" {
			parts = append(parts, fmt.Sprintf("# %s\n%s", entry.Name(), stripped))
		}
	}

	if len(parts) == 0 {
		return "", false
	}

	result := strings.Join(parts, "\n\n")
	truncated := false

	if len(result) > maxContextBytes {
		result = result[:maxContextBytes]
		truncated = true
	}

	return result, truncated
}

func stripHCLFile(src []byte, filename string) string {
	file, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return ""
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return ""
	}

	var b strings.Builder

	stripBody(&b, body, 0)

	return strings.TrimSpace(b.String())
}

func stripBody(b *strings.Builder, body *hclsyntax.Body, indent int) {
	prefix := strings.Repeat("  ", indent)

	for _, attr := range body.Attributes {
		fmt.Fprintf(b, "%s%s = %s\n", prefix, attr.Name, strippedLiteral)
	}

	for _, block := range body.Blocks {
		fmt.Fprintf(b, "%s%s", prefix, block.Type)

		for _, label := range block.Labels {
			fmt.Fprintf(b, " %q", label)
		}

		fmt.Fprintf(b, " {\n")

		if block.Body != nil {
			stripBody(b, block.Body, indent+1)
		}

		fmt.Fprintf(b, "%s}\n", prefix)
	}
}
