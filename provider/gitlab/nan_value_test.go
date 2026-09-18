// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// ".nan" is a float any YAML document may carry, and the direct conversion
// handed it to cty.NumberFloatVal, which panics on one. A single value in a
// pipeline took the whole command down with a stack trace instead of naming
// the value. The GitHub provider already declines it here.
func TestANaNIsReportedNotPanicked(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
	}{
		{
			name: "on a job attribute",
			yml:  "build:\n  script:\n    - echo hi\n  timeout: .nan\n",
		},
		{
			name: "inside a list",
			yml:  "build:\n  script:\n    - echo hi\n  parallel:\n    matrix:\n      - SIZE: [.nan]\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yml)

			if err == nil {
				t.Fatal("want a NaN reported, got no error")
			}

			if !strings.Contains(err.Error(), "NaN") {
				t.Errorf("want the message to name the NaN, got %v", err)
			}
		})
	}
}
