// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// Two variable names that sanitize to the same identifier were written under
// the same block label, so the file held "variable \"a_b\"" twice. The parse
// direction files a block's comments under its label, so the first block's
// comment was overwritten by the second and did not come back. A job and a
// template already make their labels unique.
func TestCollidingVariableNamesGetDistinctLabels(t *testing.T) {
	const yml = `stages: [build]
variables:
  # first comment
  A-B: one
  # second comment
  A_B: two
j:
  stage: build
  script: [make]
`

	hcl, back := roundtripYAML(t, yml)

	if strings.Count(hcl, `variable "a_b"`) > 1 {
		t.Errorf("two variable blocks share a label:\n%s", hcl)
	}

	for _, want := range []string{"# first comment", "# second comment"} {
		if !strings.Contains(back, want) {
			t.Errorf("%q did not survive the roundtrip:\n%s", want, back)
		}
	}
}
