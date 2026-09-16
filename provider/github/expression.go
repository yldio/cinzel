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

// validateExpressionSyntax checks that ${{ }} delimiters in a string are
// balanced and non-empty.
//
// A "}}" outside an expression is not always a mistake: a shell script holding
// JSON ends nested objects that way. The old check compared the first "}}" in
// the string against the first "${{" and refused anything where the closer came
// first, so a script with JSON and a later expression did not parse. Count the
// plain braces instead, and report a closer only when there is none for it to
// close.
func validateExpressionSyntax(path, s string) error {
	braces := 0

	for i := 0; i < len(s); {
		rest := s[i:]

		switch {
		case strings.HasPrefix(rest, "${{"):
			end := strings.Index(rest[3:], "}}")

			if end < 0 {
				return fmt.Errorf("%s: unclosed expression '${{' (missing '}}') in %q", path, s)
			}

			if strings.TrimSpace(rest[3:3+end]) == "" {
				return fmt.Errorf("%s: empty expression '${{ }}' in %q", path, s)
			}

			i += 3 + end + 2

		case strings.HasPrefix(rest, "}}"):
			// A string with no expression at all is left alone, the way it was
			// before: "}}" in a script that never interpolates is just text.
			if braces < 2 && strings.Contains(s, "${{") {
				return fmt.Errorf("%s: orphaned '}}' without matching '${{' in %q", path, s)
			}

			if braces >= 2 {
				braces -= 2
			}

			i += 2

		default:
			switch rest[0] {
			case '{':
				braces++
			case '}':
				if braces > 0 {
					braces--
				}
			}

			i++
		}
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
