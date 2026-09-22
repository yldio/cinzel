// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

// The comment on a version line is not always the "# tag" a pin left there.
// Replacing it whole deleted whatever the author had written, and reported
// success.
func TestTheAuthorsNoteSurvivesAPin(t *testing.T) {
	for _, tc := range []struct {
		name    string
		comment string
		want    string
	}{
		{name: "hash", comment: `# do not move: v5 drops node16`, want: "do not move: v5 drops node16"},
		{name: "slash", comment: `// do not move: v5 drops node16`, want: "do not move: v5 drops node16"},
		{name: "block", comment: `/* do not move */`, want: "do not move"},
		{name: "tag then note", comment: `# v4 do not move`, want: "do not move"},
		{name: "block then hash", comment: `/* keep */ # at v4`, want: "keep at v4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeHCL(t, `step "a" {
  uses {
    action  = "actions/checkout"
    version = "v4" `+tc.comment+`
  }
}
`)

			resolver := &mockResolver{shas: map[string]string{
				"actions/checkout@v4": "ffffffffffffffffffffffffffffffffffffffff",
			}}

			var buf bytes.Buffer

			if _, err := PinFile(context.Background(), path, resolver, &buf, false); err != nil {
				t.Fatal(err)
			}

			out, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			got := string(out)

			if _, err := findActionRefs(got); err != nil {
				t.Fatalf("the pinned file no longer parses: %v\n%s", err, got)
			}

			want := `version = "ffffffffffffffffffffffffffffffffffffffff" # v4 ` + tc.want

			if !strings.Contains(got, want) {
				t.Errorf("version line is not %q, in:\n%s", want, got)
			}
		})
	}
}

// A note kept beside the tag must not grow a second copy of itself, or collect
// the superseded tags, however many times the file is pinned or upgraded.
func TestPinningTwiceKeepsOneNoteAndOneTag(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "actions/checkout"
    version = "v4" # do not move: v5 drops node16
  }
}
`)

	resolver := &mockResolver{shas: map[string]string{
		"actions/checkout@v4": "ffffffffffffffffffffffffffffffffffffffff",
	}}

	var buf bytes.Buffer

	for range 3 {
		if _, err := PinFile(context.Background(), path, resolver, &buf, false); err != nil {
			t.Fatal(err)
		}
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(out)

	if n := strings.Count(got, "do not move"); n != 1 {
		t.Errorf("the note appears %d times, want 1, in:\n%s", n, got)
	}

	if n := strings.Count(got, "#"); n != 1 {
		t.Errorf("expected one comment, found %d in:\n%s", n, got)
	}
}

// Upgrading carries the note forward and names only the tag now pinned. The
// superseded one is dropped, which is the stacking the rewrite exists to stop.
func TestUpgradeKeepsTheNoteAndDropsTheOldTag(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "acme/stale"
    version = "cccccccccccccccccccccccccccccccccccccccc" # v4 see RFC-12 before bumping
  }
}
`)

	resolver := &upgraderStub{
		latest: map[string]string{"acme/stale": "v5"},
		shas:   map[string]string{"acme/stale@v5": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
	}

	var buf bytes.Buffer

	if _, err := UpgradeFile(context.Background(), path, resolver, &buf, false); err != nil {
		t.Fatal(err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got := string(out)

	want := `version = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" # v5 see RFC-12 before bumping`

	if !strings.Contains(got, want) {
		t.Errorf("version line is not %q, in:\n%s", want, got)
	}

	if strings.Contains(got, "v4") {
		t.Errorf("the superseded tag v4 is still named in:\n%s", got)
	}
}

// A comment that is only the tag a pin wrote leaves no note, so the line reads
// the way it always has.
func TestATagOnlyCommentLeavesNoNote(t *testing.T) {
	for _, comment := range []string{`# v4`, `// v4`, `/* v4 */`, `#   v4  `} {
		if note := authorNote(" "+comment, "v4"); note != "" {
			t.Errorf("authorNote(%q) = %q, want empty", comment, note)
		}
	}
}

// Only the tag naming the version on the line is a tag this tool wrote. A note
// of the author's that opens on some other version is theirs, and dropping its
// first word because it looks like a tag reads back as a sentence missing a
// word.
func TestANoteOpeningOnAnotherVersionIsKept(t *testing.T) {
	for _, tc := range []struct {
		name    string
		comment string
		version string
		want    string
	}{
		{"another version on a tag line", `# v5 drops node16, do not move`, "v4", "v5 drops node16, do not move"},
		{"the same tag on a tag line", `# v4 do not move`, "v4", "do not move"},
		{"a tag after a pin", `# v4 do not move`, strings.Repeat("a", 40), "do not move"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if note := authorNote(" "+tc.comment, tc.version); note != tc.want {
				t.Errorf("authorNote(%q, %q) = %q, want %q", tc.comment, tc.version, note, tc.want)
			}
		})
	}
}
