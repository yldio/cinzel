// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"errors"
	"fmt"
	"net/http"
)

var errNoHCLFiles = errors.New("no HCL files found in the specified path")

// validateGitHubNames checks that owner, repo, and tag contain only safe
// characters to prevent URL injection.
func validateGitHubNames(owner, repo, tag string) error {
	for _, pair := range []struct{ name, val string }{
		{"owner", owner}, {"repo", repo}, {"tag", tag},
	} {
		if !safeNamePattern.MatchString(pair.val) {
			return fmt.Errorf("invalid GitHub %s name: %q", pair.name, pair.val)
		}
	}

	return nil
}

const tokenHint = "Set GITHUB_TOKEN for authenticated requests (5000/hr vs 60/hr unauthenticated):\n  export GITHUB_TOKEN=ghp_..."

// classifyGitHubError returns a user-friendly error for GitHub API failures.
func classifyGitHubError(statusCode int, resource string, unauthenticated bool) error {
	switch statusCode {
	case http.StatusForbidden:
		if unauthenticated {
			return fmt.Errorf("GitHub API rate limit exceeded for %s.\n\n%s", resource, tokenHint)
		}

		return fmt.Errorf("GitHub API rate limit exceeded for %s. Try again later", resource)
	case http.StatusNotFound:
		return fmt.Errorf("GitHub API returned 404 for %s (action may be private or not exist)", resource)
	case http.StatusUnauthorized:
		return fmt.Errorf("GitHub API authentication failed. Check your GITHUB_TOKEN is valid")
	default:
		msg := fmt.Sprintf("GitHub API returned %d for %s", statusCode, resource)
		if unauthenticated && (statusCode == http.StatusTooManyRequests) {
			msg += "\n\n" + tokenHint
		}

		return errors.New(msg)
	}
}
