// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"strings"
	"testing"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

// Marking the sentinels in internal/hclparser was not enough to reach the
// output here: every expression refusal in this file appended the
// open-an-issue line by hand, which put it back on an error the mark had just
// taken it off. The job path, which wraps the same errors plainly, printed
// clean throughout — so the two levels disagreed about whose fault the same
// index was.
func TestAStepExpressionErrorIsTheAuthorsToFix(t *testing.T) {
	for _, tc := range []struct{ name, expr string }{
		{"an index that is not a number", `variable.envs["prod"]`},
		{"a variable nobody declared", "variable.nope"},
		{"a reference into a value", "variable.envs.prod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseSteps(t, "step \"test\" {\n  run = "+tc.expr+"\n}\n")
			if err == nil {
				t.Fatalf("parse error = nil, want %q refused", tc.expr)
			}

			if !cinzelerror.IsUserInput(err) {
				t.Errorf("error = %v, which asks the author to open an issue about their own HCL", err)
			}

			if strings.Contains(err.Error(), cinzelerror.OpenIssue) {
				t.Errorf("the open-an-issue line was written in by hand: %v", err)
			}
		})
	}
}
