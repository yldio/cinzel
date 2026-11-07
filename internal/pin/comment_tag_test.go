// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import "testing"

// A pinned line records its release in the trailing comment and nowhere else,
// so a report naming only the SHA leaves the reader to look up what it was.
// keepNote drops that word; this reads it before it goes.
func TestCommentTagReadsTheReleaseAPinnedLineNames(t *testing.T) {
	const sha = "d6db90164ac5ed86f2b6aed7e0febac5b3c0c03e"

	for _, tc := range []struct {
		name     string
		trailing string
		version  string
		want     string
	}{
		{"a pinned line", ` # v4`, sha, "v4"},
		{"a slash comment", ` // v5.0.1`, sha, "v5.0.1"},
		{"behind an author's note", ` # do not move: v4`, sha, "v4"},
		{"a block comment first", ` /* pinned */ # v4`, sha, "v4"},

		// A line still on a tag says its version in the version itself, so
		// there is nothing for the comment to add and a word that happens to
		// look like a tag is the author's.
		{"a line on a tag", ` # v4`, "v4", ""},

		{"no comment", ``, sha, ""},
		{"a note with no tag in it", ` # pinned by hand`, sha, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := commentTag(tc.trailing, tc.version); got != tc.want {
				t.Errorf("commentTag(%q, %q) = %q, want %q", tc.trailing, tc.version, got, tc.want)
			}
		})
	}
}

// The whole way through: a pinned action read out of a file carries the
// release its comment names, so the report has both ends of the move.
func TestAPinnedActionCarriesTheReleaseItsCommentNames(t *testing.T) {
	refs, err := findActionRefs(`step "deploy" {
  uses {
    action  = "actions/deploy-pages"
    version = "d6db90164ac5ed86f2b6aed7e0febac5b3c0c03e" # v4
  }
}
`)
	if err != nil {
		t.Fatal(err)
	}

	if len(refs) != 1 {
		t.Fatalf("expected one action, got %d", len(refs))
	}

	if refs[0].tagNote != "v4" {
		t.Errorf("tagNote = %q, want v4", refs[0].tagNote)
	}
}
