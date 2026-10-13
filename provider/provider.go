// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package provider

import "fmt"

// ProviderOps holds the options passed to a provider's Parse or Unparse operation.
type ProviderOps struct {
	File            string
	Directory       string
	OutputDirectory string
	Recursive       bool
	DryRun          bool
	YML             bool // use .yml extension instead of .yaml
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
