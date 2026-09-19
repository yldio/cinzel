// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"fmt"
	"strings"
)

// ValidateUsesRef checks that a step 'uses' value follows a valid format:
//   - {owner}/{repo}@{ref}
//   - {owner}/{repo}/{path}@{ref}
//   - ./local/path
//   - docker://{image}
func ValidateUsesRef(uses string) error {
	if uses == "" {
		return fmt.Errorf("uses must not be empty")
	}

	// Local action. The path is resolved against the workspace, which is the
	// repository root, so it starts there: GitHub documents "./path/to/dir"
	// and nothing else. One starting "../" points outside the checkout, where
	// the runner has nothing to find, and GitHub refuses the workflow rather
	// than running it. actionlint reads it as a remote reference and reports
	// `specifying action "../shared/action" in invalid format because ref is
	// missing`. A "../" further along the path is the author's own to resolve
	// and stays inside, so only the prefix is refused.

	if strings.HasPrefix(uses, "../") {
		return fmt.Errorf("uses %q must stay inside the repository: a local action starts with './'", uses)
	}

	if strings.HasPrefix(uses, "./") {
		return nil
	}

	// Docker action

	if strings.HasPrefix(uses, "docker://") {
		image := uses[len("docker://"):]

		if image == "" {
			return fmt.Errorf("docker uses must specify an image: %q", uses)
		}

		return nil
	}

	// Remote action: owner/repo@ref or owner/repo/path@ref
	atIdx := strings.LastIndex(uses, "@")

	if atIdx < 0 {
		return fmt.Errorf("uses %q must include a version reference (@ref, @sha, or @tag)", uses)
	}

	ref := uses[atIdx+1:]

	if ref == "" {
		return fmt.Errorf("uses %q has empty version reference after '@'", uses)
	}

	slug := uses[:atIdx]
	parts := strings.SplitN(slug, "/", 3) // owner/repo or owner/repo/path

	if len(parts) < 2 {
		return fmt.Errorf("uses %q must be in owner/repo@ref format", uses)
	}

	owner := parts[0]
	repo := parts[1]

	if owner == "" || repo == "" {
		return fmt.Errorf("uses %q has empty owner or repo name", uses)
	}

	return nil
}
