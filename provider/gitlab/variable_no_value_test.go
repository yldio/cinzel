// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// GitLab takes a variable carrying a description and no value: the name is
// listed on the manual-pipeline form with the value field left blank. Unparse
// wrote exactly that, at exit 0, and the same tool's parse direction then
// refused its own output with "must include 'name' and 'value'".
func TestVariableWithNoValueRoundtrips(t *testing.T) {
	const yml = `stages:
  - build
variables:
  A:
    description: "pick one"
    options:
      - "1"
      - "2"
a:
  stage: build
  script:
    - make
`

	_, back := roundtripYAML(t, yml)

	for _, want := range []string{"description: pick one", "options:"} {
		if !strings.Contains(back, want) {
			t.Fatalf("roundtrip lost %q:\n%s", want, back)
		}
	}

	if strings.Contains(back, "value:") {
		t.Fatalf("roundtrip invented a value:\n%s", back)
	}
}

// A variable block holding nothing at all names no value and describes none
// either, so it still has to be refused.
func TestVariableWithNothingIsRefused(t *testing.T) {
	const hcl = `variable "a" {
  name = "A"
}
`

	err := parseHCLString(t, hcl)
	if err == nil {
		t.Fatal("want an error for an empty variable block")
	}
}
