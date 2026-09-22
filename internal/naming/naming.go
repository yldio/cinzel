// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package naming

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// SanitizeIdentifier replaces non-alphanumeric characters with underscores and ensures a valid identifier.
func SanitizeIdentifier(in string) string {
	if in == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(in))

	for _, r := range in {
		if isIdentifierRune(r) {
			b.WriteRune(r)
			continue
		}

		b.WriteRune('_')
	}

	out := b.String()

	if out == "" {
		return ""
	}

	// Decoded as a rune rather than indexed as a byte. out[0] on a non-ASCII
	// digit is a UTF-8 lead byte, never a digit, so the prefix was skipped and
	// the identifier went out starting with a digit HCL refuses: the file was
	// written, the command exited 0, and parsing it back failed.
	if first, _ := utf8.DecodeRuneInString(out); unicode.IsDigit(first) {
		return "_" + out
	}

	return out
}

// isIdentifierRune reports whether r may appear in an HCL identifier.
//
// HCL's scanner carries a Unicode version of its own, and Go's tables have
// moved ahead of it: unicode.IsLetter accepts runes hclsyntax then refuses with
// "this character is not used within the language". Keeping one wrote a
// reference like "job.\u0860 alpha", which is not HCL at all — the file went out
// with exit 0 and cinzel's own parse could not read it back. ASCII is settled
// in both, so only the rest is put to hclsyntax.
func isIdentifierRune(r rune) bool {
	if r < utf8.RuneSelf {
		return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
	}

	return hclsyntax.ValidIdentifier("a" + string(r))
}

// UniqueIdentifierInSet returns base or a suffixed variant not present in the existing set.
func UniqueIdentifierInSet(base string, existing map[string]struct{}) string {
	if _, ok := existing[base]; !ok {
		return base
	}

	idx := 2

	for {
		candidate := fmt.Sprintf("%s_%d", base, idx)

		if _, ok := existing[candidate]; !ok {
			return candidate
		}

		idx++
	}
}

// ToHCLKey converts a name to an HCL-compatible key by replacing hyphens with underscores.
func ToHCLKey(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// ToYAMLKey converts a name to a YAML-compatible key by replacing underscores with hyphens.
func ToYAMLKey(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}
