// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package naming

import (
	"fmt"
	"strings"
	"unicode"
)

// SanitizeIdentifier replaces non-alphanumeric characters with underscores and ensures a valid identifier.
func SanitizeIdentifier(in string) string {
	if in == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(in))

	for _, r := range in {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
			continue
		}

		b.WriteRune('_')
	}

	out := b.String()

	if out == "" {
		return ""
	}

	if unicode.IsDigit(rune(out[0])) {
		return "_" + out
	}

	return out
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
