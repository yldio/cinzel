// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"testing"
)

// ".inf" is a float any YAML document may carry, and hclwrite has no token for
// one: it wrote "+ Inf", which is not an HCL expression at all. The file went
// out with exit 0 and cinzel's own parse then could not read it back.
// actionlint refuses ".inf" outright, reporting `invalid float value`.
func TestAnInfiniteNumberIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			name: "on a job attribute",
			yaml: "name: n\non:\n  push:\njobs:\n  a:\n    runs-on: ubuntu-latest\n" +
				"    timeout-minutes: .inf\n    steps:\n      - run: echo a\n",
		},
		{
			name: "negative, on a step env value",
			yaml: "name: n\non:\n  push:\njobs:\n  a:\n    runs-on: ubuntu-latest\n" +
				"    steps:\n      - run: echo a\n        env:\n          V: -.inf\n",
		},
		{
			name: "inside a list",
			yaml: "name: n\non:\n  push:\nenv:\n  FOO: [.inf]\njobs:\n  a:\n" +
				"    runs-on: ubuntu-latest\n    steps:\n      - run: echo a\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

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
	err, _ := unparseYAMLString(t, "name: n\non:\n  push:\njobs:\n  a:\n"+
		"    runs-on: ubuntu-latest\n    timeout-minutes: 1.5\n"+
		"    steps:\n      - run: echo a\n")

	if err != nil {
		t.Fatalf("a finite number was refused: %v", err)
	}
}
