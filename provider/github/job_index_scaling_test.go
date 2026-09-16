// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"testing"
)

// buildWorkflowJobIndex used to rebuild the set of taken job identifiers from a
// slice on every job, which made a workflow's conversion cost grow with the
// square of its job count: 8000 jobs took 322ms and allocated 1.4GB. Allocation
// is counted rather than timed so the gate does not depend on the machine.
func TestJobIndexAllocationGrowsLinearly(t *testing.T) {
	allocsFor := func(jobCount int) float64 {
		jobs := make(map[string]any, jobCount)
		order := make([]string, 0, jobCount)

		for i := range jobCount {
			name := fmt.Sprintf("job%d", i)
			jobs[name] = map[string]any{"runs-on": "ubuntu-latest"}
			order = append(order, name)
		}

		return testing.AllocsPerRun(3, func() {
			if _, _, _, err := buildWorkflowJobIndex(jobs, order, map[string]struct{}{}); err != nil {
				t.Fatal(err)
			}
		})
	}

	small, large := allocsFor(500), allocsFor(4000)

	// Eight times the jobs, so linear behaviour stays near eight times the
	// allocations. The quadratic version was over 30x here.
	if ratio := large / small; ratio > 12 {
		t.Errorf("allocations grew %.1fx for 8x the jobs, want linear growth: %.0f -> %.0f", ratio, small, large)
	}
}
