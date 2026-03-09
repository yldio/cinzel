// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

import "fmt"

// ValidateModel checks that a job's fields are consistent (e.g. uses vs runs-on).
func ValidateModel(job ValidationModel, runsOnName string) error {
	hasUses := job.Uses != ""

	if hasUses {
		if job.HasRunsOn {
			return fmt.Errorf("'uses' and '%s' cannot be defined together", runsOnName)
		}

		if job.StepCount > 0 {
			return fmt.Errorf("'uses' and 'steps' cannot be defined together")
		}

		return nil
	}

	if job.HasWith {
		return fmt.Errorf("'with' is only valid when 'uses' is set")
	}

	if job.HasSecrets {
		return fmt.Errorf("'secrets' is only valid when 'uses' is set")
	}

	if !job.HasRunsOn {
		return fmt.Errorf("'%s' is required when 'uses' is not set", runsOnName)
	}

	if job.StepCount == 0 {
		return fmt.Errorf("'steps' must contain at least one step when 'uses' is not set")
	}

	return nil
}

// ValidateNeedsReferences ensures all needs entries reference existing, non-duplicate jobs.
func ValidateNeedsReferences(needs []string, jobs map[string]ValidationModel) error {
	seen := map[string]struct{}{}

	for _, need := range needs {
		if _, exists := seen[need]; exists {
			return fmt.Errorf("contains duplicate needed job '%s'", need)
		}
		seen[need] = struct{}{}

		if _, ok := jobs[need]; !ok {
			return fmt.Errorf("cannot find needed job '%s'", need)
		}
	}

	return nil
}

// ValidateNeedsCycles detects circular dependencies in the job needs graph.
func ValidateNeedsCycles(jobs map[string]ValidationModel) error {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	color := make(map[string]int, len(jobs))

	var dfs func(id string) error
	dfs = func(id string) error {
		switch color[id] {
		case visiting:
			return fmt.Errorf("job '%s' is part of a dependency cycle", id)
		case visited:
			return nil
		}

		color[id] = visiting

		for _, dep := range jobs[id].Needs {
			if err := dfs(dep); err != nil {
				return err
			}
		}

		color[id] = visited

		return nil
	}

	for id := range jobs {
		if err := dfs(id); err != nil {
			return err
		}
	}

	return nil
}
