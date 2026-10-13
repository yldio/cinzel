// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"os"
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

	if !strings.Contains(diff, markerAdded+"INSERTED") {
		t.Errorf("the inserted line is not named\n%s", diff)
	}
}

func TestUnifiedDiffReportsAChangedLine(t *testing.T) {
	want := []byte("jobs:\n  build:\n    go-version: \"1.26\"\n    runs-on: ubuntu\n")
	got := []byte("jobs:\n  build:\n    go-version: \"1.27\"\n    runs-on: ubuntu\n")

	diff := UnifiedDiff(want, got)

	for _, expected := range []string{markerRemoved + "    go-version: \"1.26\"", markerAdded + "    go-version: \"1.27\""} {
		if !strings.Contains(diff, expected) {
			t.Errorf("diff does not carry %q\n%s", expected, diff)
		}
	}

	if strings.Contains(diff, markerRemoved+"    runs-on") {
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

	if !strings.Contains(diff, markerAdded+"name: CI") {
		t.Errorf("a file with no previous content is not all additions\n%s", diff)
	}
}

// The marker sits in its own column so the content starts in the same place
// on every line. Written against the text, "-" and "+" shift the indentation
// of exactly the lines someone is trying to read the indentation of.
func TestMarkersDoNotShiftTheContent(t *testing.T) {
	want := []byte("jobs:\n  build:\n    runs-on: ubuntu-latest\n")
	got := []byte("jobs:\n  build:\n    runs-on: ubuntu-24.04\n")

	column := -1

	for _, line := range strings.Split(UnifiedDiff(want, got), "\n") {
		if line == "" || strings.HasPrefix(line, "@@") {
			continue
		}

		at := len(line) - len(strings.TrimLeft(line[len(markerContext):], " ")) - len(markerContext)

		trimmed := strings.TrimSpace(line)
		if !strings.HasSuffix(trimmed, "runs-on: ubuntu-latest") && !strings.HasSuffix(trimmed, "runs-on: ubuntu-24.04") {
			continue
		}

		if column == -1 {
			column = at

			continue
		}

		if at != column {
			t.Errorf("the same line is indented %d on one side and %d on the other:\n%s", column, at, UnifiedDiff(want, got))
		}
	}
}

func TestColorizeMarksOnlyChangedLines(t *testing.T) {
	diff := UnifiedDiff([]byte("a\nb\nc\n"), []byte("a\nB\nc\n"))

	colored := Colorize(diff, true)

	if !strings.Contains(colored, ansiRed+markerRemoved+"b"+ansiReset) {
		t.Errorf("the removed line is not red:\n%q", colored)
	}

	if !strings.Contains(colored, ansiGreen+markerAdded+"B"+ansiReset) {
		t.Errorf("the added line is not green:\n%q", colored)
	}

	for _, line := range strings.Split(colored, "\n") {
		if strings.HasPrefix(line, "@@") && strings.Contains(line, "\x1b") {
			t.Errorf("the hunk header was coloured: %q", line)
		}

		if strings.HasPrefix(line, markerContext) && strings.Contains(line, "\x1b") {
			t.Errorf("a context line was coloured: %q", line)
		}
	}
}

func TestColorizeDisabledLeavesTheDiffAlone(t *testing.T) {
	diff := UnifiedDiff([]byte("a\n"), []byte("b\n"))

	if got := Colorize(diff, false); got != diff {
		t.Errorf("Colorize(.., false) changed the diff:\n%q", got)
	}

	if strings.Contains(Colorize(diff, false), "\x1b") {
		t.Error("an escape sequence reached a run with colour off")
	}
}

// NO_COLOR is honoured by its presence, whatever it is set to, per
// no-color.org. Set to the empty string it still means no colour.
func TestUseColorHonoursNoColor(t *testing.T) {
	for _, value := range []string{"1", "", "false", "0"} {
		t.Setenv("NO_COLOR", value)

		if UseColor(os.Stdout) {
			t.Errorf("NO_COLOR=%q did not turn colour off", value)
		}
	}
}

func TestUseColorIsOffForSomethingThatIsNotATerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	os.Unsetenv("NO_COLOR")

	if UseColor(new(strings.Builder)) {
		t.Error("a buffer was treated as a terminal")
	}

	t.Setenv("TERM", "dumb")

	if UseColor(os.Stdout) {
		t.Error("TERM=dumb did not turn colour off")
	}
}

func countChanges(diff string) (added int, removed int) {
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, markerAdded):
			added++
		case strings.HasPrefix(line, markerRemoved):
			removed++
		}
	}

	return added, removed
}
