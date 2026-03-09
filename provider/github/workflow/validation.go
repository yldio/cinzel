// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import "fmt"

// ValidateModel checks that a workflow has triggers and non-duplicate job references.
func ValidateModel(workflow ValidationModel) error {
	if !workflow.HasOn || workflow.OnCount == 0 {
		return fmt.Errorf("must define at least one trigger in 'on'")
	}

	if len(workflow.JobRefs) == 0 {
		return fmt.Errorf("must reference at least one job in 'jobs'")
	}

	seen := map[string]struct{}{}

	for _, ref := range workflow.JobRefs {
		if _, exists := seen[ref]; exists {
			return fmt.Errorf("contains duplicate job reference '%s'", ref)
		}
		seen[ref] = struct{}{}
	}

	return nil
}
