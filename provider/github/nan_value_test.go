// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// ".nan" is a float any YAML document may carry, and the direct conversion
// handed it to cty.NumberFloatVal, which panics on one. A single value in a
// workflow took the whole command down with a stack trace instead of naming
// the file and the value.
func TestANaNIsReportedNotPanicked(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			name: "on a job attribute",
			yaml: "name: n\non:\n  push:\njobs:\n  a:\n    runs-on: ubuntu-latest\n" +
				"    timeout-minutes: .nan\n    steps:\n      - run: echo a\n",
		},
		{
			name: "inside a list",
			yaml: "name: n\non:\n  push:\nenv:\n  FOO: [.nan]\njobs:\n  a:\n" +
				"    runs-on: ubuntu-latest\n    steps:\n      - run: echo a\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err, _ := unparseYAMLString(t, tc.yaml)

			if err == nil {
				t.Fatal("want a NaN reported, got no error")
			}

			if !strings.Contains(err.Error(), "NaN") {
				t.Errorf("want the message to name the NaN, got %v", err)
			}
		})
	}
}
