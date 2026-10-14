// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"path/filepath"
	"testing"
)

// Both spellings are read, so a caller setting either is understood and
// .cinzelrc.yaml — which sets the singular — keeps working.
func TestInputDirectories(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts ProviderOps
		want []string
	}{
		{
			name: "neither",
			opts: ProviderOps{},
			want: nil,
		},
		{
			name: "the singular alone",
			opts: ProviderOps{Directory: "cinzel"},
			want: []string{"cinzel"},
		},
		{
			name: "the plural alone",
			opts: ProviderOps{Directories: []string{"a", "b"}},
			want: []string{"a", "b"},
		},
		{
			name: "both, singular first",
			opts: ProviderOps{Directory: "cinzel", Directories: []string{"a"}},
			want: []string{"cinzel", "a"},
		},
		{
			name: "a repeat is one directory",
			opts: ProviderOps{Directories: []string{"a", "a"}},
			want: []string{"a"},
		},
		{
			name: "a repeat spelled differently is still one",
			opts: ProviderOps{Directories: []string{"a", "./a", "a/"}},
			want: []string{"a"},
		},
		{
			name: "an empty entry is not a directory",
			opts: ProviderOps{Directories: []string{"", "a", ""}},
			want: []string{"a"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.opts.InputDirectories()

			if len(got) != len(tc.want) {
				t.Fatalf("InputDirectories() = %v, want %v", got, tc.want)
			}

			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("InputDirectories()[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestUnderRoot(t *testing.T) {
	t.Run("no root leaves a relative path where it is", func(t *testing.T) {
		opts := ProviderOps{}

		if got := opts.UnderRoot("cinzel"); got != "cinzel" {
			t.Errorf("UnderRoot() = %q, want %q", got, "cinzel")
		}
	})

	t.Run("a root is joined on", func(t *testing.T) {
		root := filepath.Join("tmp", "x")
		opts := ProviderOps{Root: root}

		if got := opts.UnderRoot("cinzel"); got != filepath.Join(root, "cinzel") {
			t.Errorf("UnderRoot() = %q, want it under the root", got)
		}
	})

	// An absolute path names one place whatever the root says, and the flags
	// carrying one are the operator's own typing.
	//
	// Built for the platform rather than written down: "/elsewhere" is
	// absolute on Unix and relative on Windows, where absolute means a
	// volume. Spelled as a literal, this asserted that UnderRoot ignores a
	// root for a path Windows says is relative to it.
	t.Run("an absolute path is left alone", func(t *testing.T) {
		absolute, err := filepath.Abs(filepath.Join("elsewhere", "ci.hcl"))
		if err != nil {
			t.Fatal(err)
		}

		opts := ProviderOps{Root: filepath.Join("tmp", "x")}

		if got := opts.UnderRoot(absolute); got != absolute {
			t.Errorf("UnderRoot() = %q, want it untouched", got)
		}
	})

	t.Run("an empty path stays empty", func(t *testing.T) {
		opts := ProviderOps{Root: filepath.Join("tmp", "x")}

		if got := opts.UnderRoot(""); got != "" {
			t.Errorf("UnderRoot(\"\") = %q, want empty", got)
		}
	})
}

// A walk recurses, so a target inside another is already covered. Asking for
// both reported one stale file twice: harmless to a prune, which removes it
// once, and wrong in a diff, which counted it.
func TestPruneTargetsDropsWhatIsAlreadyCovered(t *testing.T) {
	for _, tc := range []struct {
		name string
		plan ParsePlan
		want int
	}{
		{
			name: "the output directory alone",
			plan: ParsePlan{OutputDirectory: "out"},
			want: 1,
		},
		{
			name: "a directory outside it is its own target",
			plan: ParsePlan{OutputDirectory: ".github/workflows", PruneDirectories: []string{"actions-a"}},
			want: 2,
		},
		{
			name: "a directory inside it is already walked",
			plan: ParsePlan{OutputDirectory: ".", PruneDirectories: []string{"actions-a"}},
			want: 1,
		},
		{
			name: "the same directory twice is one",
			plan: ParsePlan{OutputDirectory: "out", PruneDirectories: []string{"out"}},
			want: 1,
		},
		{
			name: "two outside, both kept",
			plan: ParsePlan{OutputDirectory: ".github/workflows", PruneDirectories: []string{"actions-a", "actions-b"}},
			want: 3,
		},
		{
			name: "a sibling is not covered by a name it starts with",
			plan: ParsePlan{OutputDirectory: "actions", PruneDirectories: []string{"actions-a"}},
			want: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.plan.PruneTargets(); len(got) != tc.want {
				t.Errorf("PruneTargets() = %v, want %d targets", got, tc.want)
			}
		})
	}
}

func TestChangedReportsAnythingOtherThanIdentical(t *testing.T) {
	if Changed(nil) {
		t.Error("Changed(nil) is true")
	}

	if Changed([]DiffResult{{Status: DiffIdentical}, {Status: DiffIdentical}}) {
		t.Error("every file identical read as changed")
	}

	for _, status := range []DiffStatus{DiffDiffers, DiffMissing, DiffStale} {
		if !Changed([]DiffResult{{Status: DiffIdentical}, {Status: status}}) {
			t.Errorf("a %v result read as unchanged", status)
		}
	}
}
