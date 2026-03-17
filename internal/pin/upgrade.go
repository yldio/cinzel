// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UpgradeResult holds the result of upgrading a single action.
type UpgradeResult struct {
	Action     string
	OldVersion string
	NewTag     string
	NewSHA     string
	Error      error
	WasCurrent bool // true if already on the latest version
}

// UpgradeFile reads an HCL file, checks each action for a newer release,
// and updates both the version and comment. When dryRun is true, changes
// are reported but the file is not modified.
func UpgradeFile(ctx context.Context, path string, resolver *GitHubResolver, w io.Writer, dryRun bool) ([]UpgradeResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	refs := findActionRefs(string(content))
	if len(refs) == 0 {
		return nil, nil
	}

	var results []UpgradeResult

	updated := string(content)

	for _, ref := range refs {
		parts := strings.SplitN(ref.Action, "/", 2)
		if len(parts) != 2 {
			results = append(results, UpgradeResult{
				Action: ref.Action,
				Error:  fmt.Errorf("invalid action format: %s", ref.Action),
			})

			continue
		}

		latestTag, err := resolver.LatestTag(ctx, parts[0], parts[1])
		if err != nil {
			_, _ = fmt.Fprintf(w, "warning: could not find latest version for %s: %v\n", ref.Action, err)

			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				Error:      err,
			})

			continue
		}

		// Check if already on the latest major version.
		if ref.IsTag && ref.Version == latestTag {
			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				WasCurrent: true,
			})

			continue
		}

		// Resolve the latest tag to a SHA.
		sha, err := resolver.ResolveTag(ctx, parts[0], parts[1], latestTag)
		if err != nil {
			_, _ = fmt.Fprintf(w, "warning: could not pin %s@%s: %v\n", ref.Action, latestTag, err)

			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				NewTag:     latestTag,
				Error:      err,
			})

			continue
		}

		// Replace version value.
		oldLine := fmt.Sprintf(`version = %q`, ref.Version)
		newLine := fmt.Sprintf(`version = %q`, sha)
		updated = strings.Replace(updated, oldLine, newLine, 1)

		// Add or update comment.
		updated = upsertUsesComment(updated, ref.Action, latestTag)

		_, _ = fmt.Fprintf(w, "upgraded %s: %s → %s (%s)\n", ref.Action, ref.Version, latestTag, sha[:12])

		results = append(results, UpgradeResult{
			Action:     ref.Action,
			OldVersion: ref.Version,
			NewTag:     latestTag,
			NewSHA:     sha,
		})
	}

	if !dryRun && updated != string(content) {
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			return results, fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return results, nil
}

// UpgradeDirectory upgrades all HCL files in a directory.
func UpgradeDirectory(ctx context.Context, dir string, resolver *GitHubResolver, w io.Writer, dryRun bool) ([]UpgradeResult, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var allResults []UpgradeResult

	found := false

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hcl") {
			continue
		}

		found = true
		path := filepath.Join(dir, entry.Name())

		results, err := UpgradeFile(ctx, path, resolver, w, dryRun)
		if err != nil {
			_, _ = fmt.Fprintf(w, "warning: %s: %v\n", entry.Name(), err)

			continue
		}

		allResults = append(allResults, results...)
	}

	if !found {
		return nil, errNoHCLFiles
	}

	return allResults, nil
}
