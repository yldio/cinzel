// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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

// validateExpressionSyntax checks that every ${{ in a string is closed and
// encloses something.
//
// Only an opener starts an expression. A "}}" that no "${{" opened is plain
// text to GitHub, which scans left to right the same way: "${A}}" in a shell
// script, the braces format() escapes for itself, a closer written twice by
// mistake. actionlint 1.7.12 accepts all of them, and the check that called
// them orphaned refused, in both directions, workflows GitHub runs.
func validateExpressionSyntax(path, s string) error {
	for i := 0; i < len(s); {
		rest := s[i:]

		if !strings.HasPrefix(rest, "${{") {
			i++

			continue
		}

		end := strings.Index(rest[3:], "}}")

		if end < 0 {
			return fmt.Errorf("%s: unclosed expression '${{' (missing '}}') in %q", path, s)
		}

		if strings.TrimSpace(rest[3:3+end]) == "" {
			return fmt.Errorf("%s: empty expression '${{ }}' in %q", path, s)
		}

		i += 3 + end + 2
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
