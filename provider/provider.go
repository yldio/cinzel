// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ProviderOps holds the options passed to a provider's Parse or Unparse operation.
type ProviderOps struct {
	File            string
	Directory       string
	OutputDirectory string

	// Root is the directory every other path is resolved against, so one
	// setting relocates a whole run: the input read, the output written, and
	// on GitHub the place each action's own path starts from.
	//
	// Empty means the working directory, which is where cinzel reads
	// .cinzelrc.yaml and spells its source markers from, so a run left alone
	// behaves as it always has.
	//
	// Set together with OutputDirectory it is an error rather than a
	// precedence rule: the two answer one question, and a reader should not
	// have to remember which wins.
	Root string

	Recursive bool
	DryRun    bool
	YML       bool // use .yml extension instead of .yaml
}

// ResolveRoot is the directory a run resolves its paths against.
func (opts ProviderOps) ResolveRoot() string {
	if opts.Root == "" {
		return "."
	}

	return opts.Root
}

// UnderRoot joins path onto the run's root, leaving an absolute path alone.
//
// An absolute path names one place whatever the root is, and the flags that
// carry one are the operator's own typing rather than anything read out of a
// file, so it is taken as given.
func (opts ProviderOps) UnderRoot(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(opts.ResolveRoot(), path)
}

// Provider defines the interface that each CI/CD provider must implement.
type Provider interface {
	Parse(opts ProviderOps) error
	Unparse(opts ProviderOps) error
	GetProviderName() string
	GetDescription() string
	GetParseDescription() string
	GetUnparseDescription() string
}

// PlannedOutput is one file a parse would write, and what would be in it.
type PlannedOutput struct {
	// Path is where the file goes, joined and cleaned the way the write
	// would spell it.
	Path string

	// Content is the finished bytes, markers and all, exactly as they would
	// reach disk.
	Content []byte
}

// ParsePlan is everything a parse would do to the output directory, worked out
// before any of it happens.
//
// Parse is a plan followed by a write and a prune; diff is the same plan
// followed by a read and a comparison. Keeping the two on one description is
// what stops a diff reporting something a parse would not do.
//
// CurrentOutputs and ReadSources carry what the prune needs rather than the
// prune's answer. The answer depends on what is on disk, and under parse that
// means after the writes, so asking here would be asking too early.
type ParsePlan struct {
	// OutputDirectory is the directory the prune walks.
	OutputDirectory string

	// PruneDirectories are the other directories this run wrote into, which
	// the prune walks as well.
	//
	// An action carrying a path lands outside the output directory, and the
	// prune only ever looked there: renaming one left the old action.yml on
	// disk, live and runnable, which is the bug the marker was added to stop.
	//
	// Only directories this run actually produced an action in. Walking the
	// whole root instead would read .git and node_modules on every parse, to
	// catch the one case this misses — an action moved from one folder to
	// another, where the old folder is no longer anything the run touches.
	PruneDirectories []string

	// Outputs are the files to write, in the order the run produces them.
	Outputs []PlannedOutput

	// CurrentOutputs names every file this run produces, cleaned, as the
	// prune's "leave these alone" set. Empty for a provider that does not
	// prune.
	CurrentOutputs map[string]struct{}

	// ReadSources names the HCL files the run read, spelled by
	// fsutil.SourceKey. Nil where the provider records no source.
	ReadSources map[string]struct{}
}

// Print writes the plan the way a dry run shows it: each file named, then its
// contents.
func (plan ParsePlan) Print() {
	for _, out := range plan.Outputs {
		fmt.Printf("# file: %s\n", out.Path)
		fmt.Println(string(out.Content))
	}
}

// DiffStatus is what a diff found for one path.
type DiffStatus int

const (
	// DiffIdentical means the file on disk is what a parse would write.
	DiffIdentical DiffStatus = iota

	// DiffDiffers means the file is there and says something else.
	DiffDiffers

	// DiffMissing means a parse would create the file.
	DiffMissing

	// DiffStale means the file is cinzel's and a parse would delete it,
	// which is what a renamed workflow leaves behind.
	DiffStale
)

// DiffResult is one path and what a parse would do to it.
type DiffResult struct {
	Path   string
	Status DiffStatus

	// Unified is the diff from what is on disk to what would be written,
	// set only for DiffDiffers.
	Unified string
}

// PruneTargets are the directories a prune or a diff has to look in: the
// output directory, and each one an action was written to.
//
// A directory already covered by another is dropped. The walk recurses, so
// an output directory of "." contains every action directory under it, and
// asking for both reported one stale file twice — harmless to a prune, which
// removes it once, and wrong in a diff, which counted it twice and said so.
func (plan ParsePlan) PruneTargets() []string {
	targets := append([]string{plan.OutputDirectory}, plan.PruneDirectories...)

	// Shortest first, so a directory is compared against the one that would
	// contain it rather than the other way round.
	sort.Slice(targets, func(i, j int) bool { return len(targets[i]) < len(targets[j]) })

	kept := make([]string, 0, len(targets))

	for _, target := range targets {
		if !coveredBy(target, kept) {
			kept = append(kept, target)
		}
	}

	return kept
}

// coveredBy reports whether a walk of one of the directories in covers target.
func coveredBy(target string, in []string) bool {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false
	}

	for _, candidate := range in {
		absCandidate, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		if absTarget == absCandidate {
			return true
		}

		if strings.HasPrefix(absTarget, absCandidate+string(filepath.Separator)) {
			return true
		}
	}

	return false
}
