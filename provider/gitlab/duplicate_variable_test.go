// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// Two variable blocks naming the same environment variable wrote the last one
// and dropped the rest without a word, so a pipeline shipped with a token
// nobody in the file meant to set. A job and a template already refuse the
// same input.
func TestDuplicateVariableNameIsRefused(t *testing.T) {
	const hcl = `stages = ["build"]

variable "a" {
  name  = "TOKEN"
  value = "one"
}

variable "b" {
  name  = "TOKEN"
  value = "two"
}

job "j" {
  stage  = "build"
  script = ["make"]
}
`

	err := parseHCLString(t, hcl)
	if err == nil {
		t.Fatal("want a duplicate variable error")
	}

	if !strings.Contains(err.Error(), "duplicate variable name 'TOKEN'") {
		t.Fatalf("error does not name the variable: %v", err)
	}
}
