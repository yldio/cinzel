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

// Upgrader extends Resolver with the ability to find the latest release tag.
type Upgrader interface {
	Resolver
	LatestTag(ctx context.Context, owner, repo string) (string, error)
}

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
func UpgradeFile(ctx context.Context, path string, resolver Upgrader, w io.Writer, dryRun bool) ([]UpgradeResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	refs, err := findActionRefs(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse action refs in %s: %w", path, err)
	}

	if len(refs) == 0 {
		return nil, nil
	}

	var results []UpgradeResult

	var edits []versionEdit

	for _, ref := range refs {
		// Named as it happens, for the reason given in PinFile.
		owner, repo, ok := splitAction(ref.Action)
		if !ok {
			err := errNotRemoteAction(ref.Action)

			reportf(w, "warning: could not upgrade %s: %v\n", ref.Action, err)

			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				Error:      err,
			})

			continue
		}

		latestTag, err := resolver.LatestTag(ctx, owner, repo)
		if err != nil {
			reportf(w, "warning: could not find latest version for %s: %v\n", ref.Action, err)

			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				Error:      err,
			})

			continue
		}

		// Resolve the latest tag to a SHA.
		sha, err := resolver.ResolveTag(ctx, owner, repo, latestTag)

		// See PinFile: a response with no "sha" decodes to "" and no error,
		// and writing it out reports an upgrade that did not happen.
		if err == nil && !isCommitSHA(sha) {
			err = errShortSHA(sha)
		}

		if err != nil {
			reportf(w, "warning: could not pin %s@%s: %v\n", ref.Action, latestTag, err)

			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				NewTag:     latestTag,
				Error:      err,
			})

			continue
		}

		// Already on the latest version — compare by tag or SHA.
		if (ref.IsTag && ref.Version == latestTag) || (!ref.IsTag && ref.Version == sha) {
			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				WasCurrent: true,
			})

			continue
		}

		// Written at this action's own offsets, for the reason given in
		// PinFile. An upgrade reaches it more easily still: an action already
		// on the latest tag is skipped, and that is enough on its own to leave
		// an earlier line matching, no failed request needed.
		edits = append(edits, versionEdit{
			start: ref.start,
			end:   ref.end,
			text:  versionLine(sha, latestTag, ref.note),
		})

		reportf(w, "upgraded %s: %s → %s (%s)\n", ref.Action, ref.Version, latestTag, shortSHA(sha))

		results = append(results, UpgradeResult{
			Action:     ref.Action,
			OldVersion: ref.Version,
			NewTag:     latestTag,
			NewSHA:     sha,
		})
	}

	updated := applyVersionEdits(string(content), edits)

	if !dryRun && updated != string(content) {
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			return results, fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return results, nil
}

// UpgradeDirectory upgrades all HCL files in a directory.
func UpgradeDirectory(ctx context.Context, dir string, resolver Upgrader, w io.Writer, dryRun bool) ([]UpgradeResult, error) {
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
			reportf(w, "warning: %s: %v\n", entry.Name(), err)

			// Carried as a result for the same reason as in PinDirectory: the
			// summary counts results, so a file that could not be read at all
			// was reported as no failure.
			allResults = append(allResults, UpgradeResult{Action: path, Error: err})

			continue
		}

		allResults = append(allResults, results...)
	}

	if !found {
		return nil, errNoHCLFiles
	}

	return allResults, nil
}
