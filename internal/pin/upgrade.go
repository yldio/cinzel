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
	"time"
)

// Upgrader extends Resolver with the ability to find the latest release.
type Upgrader interface {
	Resolver
	LatestRelease(ctx context.Context, owner, repo string) (Release, error)
}

// Release is a repository's latest release: the tag to move to, and when it
// was published.
type Release struct {
	Tag       string
	Published time.Time
}

// Cooldown holds back a release until it has been out for a while.
//
// An upgrade run takes whatever the newest tag is, which means a release
// published minutes ago lands in CI before anyone has used it. A tag can also
// be moved or pulled after publication, and the window is the time in which
// that happens. Waiting is the whole mitigation, so the policy is one
// duration.
//
// The zero value waits for nothing, which is what upgrade did before this
// existed.
type Cooldown struct {
	// For is how long a release has to have been published before an upgrade
	// takes it.
	For time.Duration

	// Now is read in tests. Nil means time.Now.
	Now func() time.Time
}

// age is how long ago a release was published, by the clock the policy reads.
func (c Cooldown) age(r Release) time.Duration {
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}

	return now().Sub(r.Published)
}

// holds reports whether a release is still inside the cooldown, and how much
// of it is left.
//
// A release with no publication date is taken rather than held. GitHub leaves
// published_at null on a draft, and an unknown date is not evidence that
// something is new — holding on it would stall an upgrade for a reason the
// run cannot explain.
func (c Cooldown) holds(r Release) (bool, time.Duration) {
	if c.For <= 0 || r.Published.IsZero() {
		return false, 0
	}

	age := c.age(r)
	if age >= c.For {
		return false, 0
	}

	return true, c.For - age
}

// roundDuration writes a duration the way someone reading a CI log reads one:
// in days once it is past a day, and never to the nanosecond.
func roundDuration(d time.Duration) string {
	if d >= 48*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}

	if d >= time.Hour {
		return d.Round(time.Hour).String()
	}

	return d.Round(time.Minute).String()
}

// UpgradeResult holds the result of upgrading a single action.
type UpgradeResult struct {
	Action string

	// OldVersion is what the line held, and OldTag the release that was, when
	// the line was pinned and its comment named one. A pinned line records
	// the release nowhere else, so a report carrying only the SHA leaves the
	// reader to look up what it was.
	OldVersion string
	OldTag     string

	NewTag     string
	NewSHA     string
	Error      error
	WasCurrent bool // true if already on the latest version

	// HeldFor is how much of the cooldown a release still had left when the
	// run reached it. Non-zero means the upgrade was available and declined,
	// which is neither current nor failed.
	HeldFor time.Duration

	// Skipped marks an action a targeted run did not look at. It is neither
	// current nor failed: the run has no opinion about it.
	Skipped bool

	// Published is when the release the run looked at was published, and
	// EligibleAt when it leaves the cooldown. Both are carried rather than
	// only the remaining duration: a duration is read against the clock at
	// the moment of the run, so a report written now and read tomorrow says
	// something that is no longer true.
	Published  time.Time
	EligibleAt time.Time
}

// UpgradeOptions is how a run is to be carried out.
//
// A struct rather than three more parameters: the signature was at six and
// two of the new ones are strings sitting beside the path, where swapping a
// pair still compiles and the mistake shows up as a file that was not found.
type UpgradeOptions struct {
	// DryRun reports what would change without writing it.
	DryRun bool

	// Cooldown holds a release back until it has been published a while.
	Cooldown Cooldown

	// Only, when set, is the one action to consider. Every other action in
	// the file is left alone and reported as skipped, so a run that names an
	// action nobody wrote can be told from one that found nothing to do.
	Only string
}

// UpgradeFile reads an HCL file, checks each action for a newer release,
// and updates both the version and comment. When opts.DryRun is true, changes
// are reported but the file is not modified.
func UpgradeFile(ctx context.Context, path string, resolver Upgrader, w io.Writer, opts UpgradeOptions) ([]UpgradeResult, error) {
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
		// Before anything is resolved, so a targeted run makes one request
		// rather than one per action in the file.
		if opts.Only != "" && ref.Action != opts.Only {
			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				OldTag:     ref.tagNote,
				Skipped:    true,
			})

			continue
		}

		// Named as it happens, for the reason given in PinFile.
		owner, repo, ok := splitAction(ref.Action)
		if !ok {
			err := errNotRemoteAction(ref.Action)

			reportf(w, "warning: could not upgrade %s: %v\n", ref.Action, err)

			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				OldTag:     ref.tagNote,
				Error:      err,
			})

			continue
		}

		latest, err := resolver.LatestRelease(ctx, owner, repo)
		if err != nil {
			reportf(w, "warning: could not find latest version for %s: %v\n", ref.Action, err)

			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				OldTag:     ref.tagNote,
				Error:      err,
			})

			continue
		}

		latestTag := latest.Tag

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
				OldTag:     ref.tagNote,
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
				OldTag:     ref.tagNote,
				WasCurrent: true,
			})

			continue
		}

		// There is an upgrade, and the policy says not yet. Checked after
		// "already current" so an action sitting on the newest tag is not
		// reported as held for as long as the window lasts, which would read
		// as an upgrade waiting that is not.
		if held, left := opts.Cooldown.holds(latest); held {
			// Three numbers with no nouns on them — "published 65d ago, 24d
			// left" — leave the reader working out what the window was and
			// what the wait ends in, and the two floored figures do not add
			// up to it either. The date is the thing being waited for, so it
			// is what the line says.
			reportf(w, "held %s: %s → %s\n  published %s, inside the %s cooldown\n  upgrades on or after %s\n",
				ref.Action, ref.Version, latestTag,
				latest.Published.UTC().Format(time.DateOnly),
				roundDuration(opts.Cooldown.For),
				latest.Published.Add(opts.Cooldown.For).UTC().Format(time.DateOnly))

			// The SHA is already resolved above, so carrying it costs
			// nothing and every entry reports both ends of the move the
			// same way, held or not.
			results = append(results, UpgradeResult{
				Action:     ref.Action,
				OldVersion: ref.Version,
				OldTag:     ref.tagNote,
				NewTag:     latestTag,
				NewSHA:     sha,
				HeldFor:    left,
				Published:  latest.Published,
				EligibleAt: latest.Published.Add(opts.Cooldown.For),
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

		reportf(w, "%s %s: %s → %s (%s)\n", upgradedVerb(opts.DryRun), ref.Action, ref.Version, latestTag, shortSHA(sha))

		results = append(results, UpgradeResult{
			Action:     ref.Action,
			OldVersion: ref.Version,
			OldTag:     ref.tagNote,
			NewTag:     latestTag,
			NewSHA:     sha,
		})
	}

	updated := applyVersionEdits(string(content), edits)

	if !opts.DryRun && updated != string(content) {
		if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
			return results, fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return results, nil
}

// upgradedVerb is resolvedVerb for the upgrade side.
func upgradedVerb(dryRun bool) string {
	if dryRun {
		return "would upgrade"
	}

	return "upgraded"
}

// UpgradeDirectory upgrades all HCL files in a directory.
func UpgradeDirectory(ctx context.Context, dir string, resolver Upgrader, w io.Writer, opts UpgradeOptions) ([]UpgradeResult, error) {
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

		results, err := UpgradeFile(ctx, path, resolver, w, opts)
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
