// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yldio/cinzel/internal/fsutil"
)

// Diff reports what a parse would change, without changing anything.
//
// Each planned output is read back from where it would be written: absent is
// a file the parse would create, present and equal is nothing to do, and
// present and different carries the diff. Then the prune is asked what it
// would remove, which is how a renamed workflow's leftover is told apart from
// a file whose contents drifted — to a plain directory comparison the two
// look the same, and the leftover never clears.
//
// Results come back in the order the plan produces them, with the stale files
// last. An empty slice means the committed tree is exactly what this HCL
// produces.
func (plan ParsePlan) Diff(providerName string) ([]DiffResult, error) {
	results := make([]DiffResult, 0, len(plan.Outputs))

	for _, out := range plan.Outputs {
		onDisk, err := os.ReadFile(out.Path)

		if os.IsNotExist(err) {
			results = append(results, DiffResult{Path: out.Path, Status: DiffMissing})

			continue
		}

		if err != nil {
			return nil, err
		}

		if fsutil.SameContent(onDisk, out.Content) {
			results = append(results, DiffResult{Path: out.Path, Status: DiffIdentical})

			continue
		}

		results = append(results, DiffResult{
			Path:    out.Path,
			Status:  DiffDiffers,
			Unified: fsutil.UnifiedDiff(onDisk, out.Content),
		})
	}

	// A provider that does not prune records no current outputs, and asking
	// anyway would read every generated file in the directory as stale.
	if len(plan.CurrentOutputs) == 0 {
		return results, nil
	}

	stale := []string{}

	for _, dir := range plan.PruneTargets() {
		found, err := fsutil.StaleGeneratedYAML(dir, plan.CurrentOutputs, providerName, plan.ReadSources)
		if err != nil {
			return nil, err
		}

		stale = append(stale, found...)
	}

	for _, path := range stale {
		// The walk returns absolute paths, because the prune compares them
		// against the output directory. Reported beside the planned outputs
		// they would be the one absolute path in the list, so they are
		// spelled the way the rest of the report spells a file.
		results = append(results, DiffResult{Path: relativeToWorkingDir(path), Status: DiffStale})
	}

	return results, nil
}

// Changed reports whether any result is something a parse would do.
func Changed(results []DiffResult) bool {
	for _, result := range results {
		if result.Status != DiffIdentical {
			return true
		}
	}

	return false
}

// relativeToWorkingDir spells a path the way the planned outputs are spelled,
// falling back to what it was given when that cannot be worked out.
func relativeToWorkingDir(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}

	rel, err := filepath.Rel(cwd, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}

	return rel
}
