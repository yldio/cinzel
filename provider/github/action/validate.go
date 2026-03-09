// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

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

	// Local action

	if strings.HasPrefix(uses, "./") || strings.HasPrefix(uses, "../") {

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
