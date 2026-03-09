// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"fmt"
	"strings"
)

// validateExpressions walks the workflow map and checks all string values
// for well-formed ${{ }} expressions (balanced delimiters, non-empty body).
func validateExpressions(workflow map[string]any) error {

	return walkStrings(workflow, "", func(path, value string) error {

		return validateExpressionSyntax(path, value)
	})
}

// validateExpressionSyntax checks that ${{ }} delimiters in a string are
// balanced and non-empty.
func validateExpressionSyntax(path, s string) error {
	rest := s

	for {
		idx := strings.Index(rest, "${{")

		if idx < 0 {
			break
		}

		after := rest[idx+3:]
		end := strings.Index(after, "}}")

		if end < 0 {

			return fmt.Errorf("%s: unclosed expression '${{' (missing '}}') in %q", path, s)
		}

		body := strings.TrimSpace(after[:end])

		if body == "" {

			return fmt.Errorf("%s: empty expression '${{ }}' in %q", path, s)
		}

		rest = after[end+2:]
	}

	// Check for orphaned }} without opening ${{
	temp := s

	for {
		openIdx := strings.Index(temp, "${{")
		closeIdx := strings.Index(temp, "}}")

		if closeIdx < 0 {
			break
		}

		if openIdx < 0 || closeIdx < openIdx {
			// }} appears before any ${{ — could be a false positive in non-expression contexts.
			// Only flag if the string contains at least one ${{ somewhere.

			if strings.Contains(s, "${{") {

				return fmt.Errorf("%s: orphaned '}}' without matching '${{' in %q", path, s)
			}
			break
		}
		// Skip past this matched pair.
		end := strings.Index(temp[openIdx+3:], "}}")

		if end < 0 {
			break
		}
		temp = temp[openIdx+3+end+2:]
	}

	return nil
}

// walkStrings recursively visits all string values in a nested map/slice structure,
// calling fn with the dotted path and string value.
func walkStrings(v any, path string, fn func(path, value string) error) error {
	switch val := v.(type) {
	case string:
		return fn(path, val)
	case map[string]any:
		for key, child := range val {
			childPath := key

			if path != "" {
				childPath = path + "." + key
			}

			if err := walkStrings(child, childPath, fn); err != nil {

				return err
			}
		}
	case []any:
		for i, child := range val {
			childPath := fmt.Sprintf("%s[%d]", path, i)

			if err := walkStrings(child, childPath, fn); err != nil {

				return err
			}
		}
	}

	return nil
}
