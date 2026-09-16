// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShortSHA(t *testing.T) {
	full := "abc123def456abc123def456abc123def456abc1"

	for _, tc := range []struct {
		name string
		sha  string
		want string
	}{
		{name: "a full sha is abbreviated", sha: full, want: "abc123def456"},
		{name: "exactly twelve is left whole", sha: "abc123def456", want: "abc123def456"},
		{name: "a short one is left whole", sha: "abc123", want: "abc123"},
		{name: "empty", sha: "", want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := shortSHA(tc.sha); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsCommitSHA(t *testing.T) {
	for _, tc := range []struct {
		name string
		sha  string
		want bool
	}{
		{name: "a full lower-case sha", sha: strings.Repeat("a", 40), want: true},
		{name: "upper case is still hex", sha: strings.Repeat("A", 40), want: true},
		{name: "empty", sha: "", want: false},
		{name: "too short", sha: "abc123", want: false},
		{name: "too long", sha: strings.Repeat("a", 41), want: false},
		{name: "right length, not hex", sha: strings.Repeat("z", 40), want: false},
		{name: "a tag, not a sha", sha: "v4", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCommitSHA(tc.sha); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// A resolve that comes back with something short used to be sliced to twelve,
// which panicked mid-run after earlier lines had already been rewritten. It is
// now refused, so the file keeps its tag rather than gaining a version = "".
func TestAShortSHAIsRefusedNotWritten(t *testing.T) {
	const content = `step "c" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

	for _, tc := range []struct {
		name string
		sha  string
	}{
		{name: "empty", sha: ""},
		{name: "too short to slice", sha: "abc123"},
		{name: "right length, not hex", sha: strings.Repeat("z", 40)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "a.hcl")

			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}

			resolver := &mockResolver{shas: map[string]string{"actions/checkout@v4": tc.sha}}

			var buf bytes.Buffer

			results, err := PinFile(context.Background(), path, resolver, &buf, false)
			if err != nil {
				t.Fatal(err)
			}

			if len(results) != 1 || results[0].Error == nil {
				t.Fatalf("want one failed result, got %+v", results)
			}

			if !strings.Contains(buf.String(), "not a commit SHA") {
				t.Errorf("want the reason on stdout, got %q", buf.String())
			}

			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			if string(after) != content {
				t.Errorf("file was rewritten:\n%s", after)
			}
		})
	}
}

// UpgradeFile has the same resolve, and the same consequence: a bad response
// used to panic on the progress line, and after the length check alone it
// would have written the new tag's comment beside an empty version.
func TestUpgradeRefusesAShortSHA(t *testing.T) {
	const content = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

	path := filepath.Join(t.TempDir(), "steps.hcl")

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	resolver := &stubUpgrader{
		latestTags: map[string]string{"actions/checkout": "v5"},
		shas:       map[string]string{"actions/checkout@v5": "abc123"},
	}

	var buf bytes.Buffer

	results, err := UpgradeFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 || results[0].Error == nil {
		t.Fatalf("want one failed result, got %+v", results)
	}

	if !strings.Contains(buf.String(), "not a commit SHA") {
		t.Errorf("want the reason on stdout, got %q", buf.String())
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(after) != content {
		t.Errorf("file was rewritten:\n%s", after)
	}
}
