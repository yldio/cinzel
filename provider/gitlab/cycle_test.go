// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A cycle is reported so the author can break it, and breaking it means
// knowing where it runs. The message named no job at all, and the walk started
// from a map range, so which cycle a pipeline holding two of them reported
// changed from one run to the next.
func TestDependsOnCycleNamesTheJobsItRunsThrough(t *testing.T) {
	const hcl = `stages = ["build"]

job "a" {
  stage      = "build"
  script     = ["make"]
  depends_on = [job.b]
}

job "b" {
  stage      = "build"
  script     = ["make"]
  depends_on = [job.a]
}
`

	first := ""

	for range 20 {
		err := parseHCLString(t, hcl)
		if err == nil {
			t.Fatal("want a cycle error")
		}

		got := err.Error()

		if !strings.Contains(got, "'a' -> 'b' -> 'a'") {
			t.Fatalf("error does not trace the cycle: %v", err)
		}

		if first == "" {
			first = got
		}

		if got != first {
			t.Fatalf("cycle error changes between runs:\n%s\n%s", first, got)
		}
	}
}
