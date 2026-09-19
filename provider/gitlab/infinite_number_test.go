// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"testing"
)

// ".inf" is a float any YAML document may carry, and hclwrite has no token for
// one: it wrote "+ Inf", which is not an HCL expression at all. The file went
// out with exit 0 and cinzel's own parse then could not read it back. The
// GitHub provider declines it the same way.
func TestAnInfiniteNumberIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
	}{
		{
			name: "on a job attribute",
			yml:  "build:\n  script:\n    - echo hi\n  timeout: .inf\n",
		},
		{
			name: "negative, on a variable",
			yml:  "build:\n  script:\n    - echo hi\n  variables:\n    V: -.inf\n",
		},
		{
			name: "inside a list",
			yml:  "build:\n  script:\n    - echo hi\n  parallel:\n    matrix:\n      - SIZE: [.inf]\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yml)

			if err == nil {
				t.Fatal("an infinity was written to HCL as '+ Inf'")
			}

			if !errors.Is(err, errInfiniteNumber) {
				t.Fatalf("the error does not name the infinity: %v", err)
			}
		})
	}
}

// A float that is a number keeps going through.
func TestAFiniteNumberIsKept(t *testing.T) {
	if err := unparseYAMLString(t, "build:\n  script:\n    - echo hi\n  timeout: 1.5\n"); err != nil {
		t.Fatalf("a finite number was refused: %v", err)
	}
}
