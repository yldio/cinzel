// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	maxContextTokens = 8000
	bytesPerToken    = 4
	maxContextBytes  = maxContextTokens * bytesPerToken
	strippedLiteral  = `"..."`
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
		result = truncateAtNewline(result, maxContextBytes)
		truncated = true
	}

	return result, truncated
}

// truncateAtNewline truncates s to at most maxLen bytes, cutting at the last
// newline before the limit to avoid splitting mid-line or mid-rune.
func truncateAtNewline(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	cut := s[:maxLen]

	if i := strings.LastIndex(cut, "\n"); i > 0 {
		return cut[:i]
	}

	return cut
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

	// Sort attribute names for deterministic output.
	attrNames := make([]string, 0, len(body.Attributes))
	for name := range body.Attributes {
		attrNames = append(attrNames, name)
	}

	sort.Strings(attrNames)

	for _, name := range attrNames {
		fmt.Fprintf(b, "%s%s = %s\n", prefix, name, strippedLiteral)
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
