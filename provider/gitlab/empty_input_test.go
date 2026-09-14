// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"testing"
)

// An input declaring nothing used to be written out as a "{}" document.
func TestEmptyInputIsRejected(t *testing.T) {
	for name, hcl := range map[string]string{
		"empty file":      "",
		"only a comment":  "// nothing here\n",
		"only whitespace": "\n\n\n",
	} {
		t.Run(name, func(t *testing.T) {
			if err := parseHCLString(t, hcl); !errors.Is(err, errNoDefinitions) {
				t.Fatalf("expected errNoDefinitions, got %v", err)
			}
		})
	}
}

// TestRealInputStillParses guards the fix.
func TestRealInputStillParses(t *testing.T) {
	if err := parseHCLString(t, `job "build" {
  stage  = "build"
  script = ["make"]
}
`); err != nil {
		t.Fatalf("a real pipeline should parse: %v", err)
	}
}
