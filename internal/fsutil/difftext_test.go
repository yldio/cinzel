// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"strings"
	"testing"
)

func TestSameContentIgnoresLineEndings(t *testing.T) {
	crlf := []byte("name: CI\r\njobs:\r\n  build: {}\r\n")
	lf := []byte("name: CI\njobs:\n  build: {}\n")

	if !SameContent(crlf, lf) {
		t.Error("a CRLF checkout of the same file read as different content")
	}

	if SameContent(lf, []byte("name: CD\njobs:\n  build: {}\n")) {
		t.Error("a changed line read as the same content")
	}
}

// The reason this is an LCS and not a walk down both sides at once. Inserting
// a line shifts every line after it, and comparing line i against line i
// reports all of them as changed: the one line that moved is then somewhere
// in a diff the length of the file.
func TestInsertionDoesNotRewriteTheRestOfTheFile(t *testing.T) {
	want := []byte("a\nb\nc\nd\ne\nf\ng\nh\n")
	got := []byte("a\nINSERTED\nb\nc\nd\ne\nf\ng\nh\n")

	diff := UnifiedDiff(want, got)

	added, removed := countChanges(diff)

	if added != 1 {
		t.Errorf("added %d lines, want 1\n%s", added, diff)
	}

	if removed != 0 {
		t.Errorf("removed %d lines, want 0\n%s", removed, diff)
	}

	if !strings.Contains(diff, "+INSERTED") {
		t.Errorf("the inserted line is not named\n%s", diff)
	}
}

func TestUnifiedDiffReportsAChangedLine(t *testing.T) {
	want := []byte("jobs:\n  build:\n    go-version: \"1.26\"\n    runs-on: ubuntu\n")
	got := []byte("jobs:\n  build:\n    go-version: \"1.27\"\n    runs-on: ubuntu\n")

	diff := UnifiedDiff(want, got)

	for _, expected := range []string{"-    go-version: \"1.26\"", "+    go-version: \"1.27\""} {
		if !strings.Contains(diff, expected) {
			t.Errorf("diff does not carry %q\n%s", expected, diff)
		}
	}

	if strings.Contains(diff, "-    runs-on") {
		t.Errorf("an unchanged line was reported as removed\n%s", diff)
	}
}

func TestUnifiedDiffOfIdenticalContentIsEmpty(t *testing.T) {
	same := []byte("name: CI\njobs: {}\n")

	if diff := UnifiedDiff(same, same); diff != "" {
		t.Errorf("UnifiedDiff() = %q, want empty", diff)
	}
}

// A long file with one change at each end produces two hunks rather than one
// spanning everything between them.
func TestDistantChangesAreSeparateHunks(t *testing.T) {
	var want, got strings.Builder

	want.WriteString("first\n")
	got.WriteString("FIRST\n")

	for range 30 {
		want.WriteString("middle\n")
		got.WriteString("middle\n")
	}

	want.WriteString("last\n")
	got.WriteString("LAST\n")

	diff := UnifiedDiff([]byte(want.String()), []byte(got.String()))

	if hunks := strings.Count(diff, "@@ -"); hunks != 2 {
		t.Errorf("got %d hunks, want 2\n%s", hunks, diff)
	}

	if strings.Count(diff, "middle") > diffContextLines*2 {
		t.Errorf("unchanged lines between the hunks were printed\n%s", diff)
	}
}

func TestUnifiedDiffHandlesAnEmptySide(t *testing.T) {
	diff := UnifiedDiff(nil, []byte("name: CI\n"))

	if !strings.Contains(diff, "+name: CI") {
		t.Errorf("a file with no previous content is not all additions\n%s", diff)
	}
}

func countChanges(diff string) (added int, removed int) {
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			added++
		case strings.HasPrefix(line, "-"):
			removed++
		}
	}

	return added, removed
}
