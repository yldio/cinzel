// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// markedFrom returns generated YAML recording source as the file it came from.
func markedFrom(source string) []byte {
	return PrependGeneratedMarker([]byte("name: ci\n"), "github", source)
}

// TestPruneKeepsAFileFromASourceTheRunDidNotRead is the narrowing bug: a run
// given one HCL file used to delete the output of every other one, because the
// marker said cinzel wrote the file and nothing about which file asked for it.
func TestPruneKeepsAFileFromASourceTheRunDidNotRead(t *testing.T) {
	dir := t.TempDir()
	other := filepath.Join(dir, "ci.yaml")

	if err := os.WriteFile(other, markedFrom("src/ci.hcl"), 0o644); err != nil {
		t.Fatal(err)
	}

	read := map[string]struct{}{"src/cd.hcl": {}}

	if err := PruneStaleGeneratedYAML(dir, map[string]struct{}{}, "github", read); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(other); err != nil {
		t.Fatalf("the prune deleted an output of an HCL file it never read: %v", err)
	}
}

// TestPruneRemovesAFileFromASourceTheRunRead is the rename: the source was read
// and no longer writes that output, so the leftover is this run's to clean up.
func TestPruneRemovesAFileFromASourceTheRunRead(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old-name.yaml")

	if err := os.WriteFile(old, markedFrom("src/ci.hcl"), 0o644); err != nil {
		t.Fatal(err)
	}

	read := map[string]struct{}{"src/ci.hcl": {}}

	if err := PruneStaleGeneratedYAML(dir, map[string]struct{}{}, "github", read); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("expected the renamed file's old output to be removed, got %v", err)
	}
}

// TestPruneRemovesAFileWithNoSourceLine covers every file generated before the
// source line existed. Absent is not "belongs to nobody": it is no information,
// and the answer without information is the one the prune gave before.
func TestPruneRemovesAFileWithNoSourceLine(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "stale.yaml")

	if err := os.WriteFile(stale, []byte("# generated-by: cinzel\n# cinzel-provider: github\nname: ci\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	read := map[string]struct{}{"src/ci.hcl": {}}

	if err := PruneStaleGeneratedYAML(dir, map[string]struct{}{}, "github", read); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("expected a file written before the source line to be removed, got %v", err)
	}
}

// TestWithoutGeneratedMarkerStripsTheSourceLine guards the roundtrip: the
// source line names an HCL file, and unparsing that YAML writes to the file it
// names, so a line left in lands as a comment in its own source.
func TestWithoutGeneratedMarkerStripsTheSourceLine(t *testing.T) {
	comment := "# generated-by: cinzel\n# cinzel-provider: github\n# cinzel-source: src/ci.hcl"

	if got := WithoutGeneratedMarker(comment); got != "" {
		t.Fatalf("want the markers gone, got %q", got)
	}

	kept := WithoutGeneratedMarker(comment + "\n# what an author wrote")

	if kept != "# what an author wrote" {
		t.Fatalf("want only the author's comment, got %q", kept)
	}
}

// TestSourceKeySpellsOnePathOneWay pins that the two sides of the ownership
// question agree: a path recorded through one spelling has to match the same
// file reached through another, or a rename's leftover is never recognised.
func TestSourceKeySpellsOnePathOneWay(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	nested := filepath.Join(dir, "src")

	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	want := SourceKey(filepath.Join(nested, "ci.hcl"))

	if want != "src/ci.hcl" {
		t.Fatalf("want a path relative to the working directory, got %q", want)
	}

	for _, spelling := range []string{"src/ci.hcl", "./src/ci.hcl", "src/../src/ci.hcl"} {
		if got := SourceKey(spelling); got != want {
			t.Errorf("SourceKey(%q) = %q, want %q", spelling, got, want)
		}
	}
}

// TestSourceKeyOutsideTheWorkingDirectoryIsAbsolute covers a file read from
// somewhere the run is not standing. Relative, it would come back as a run of
// "..", and how many depends on the caller's depth, so the same file read from
// two directories would be two different files to the prune.
func TestSourceKeyOutsideTheWorkingDirectoryIsAbsolute(t *testing.T) {
	outside := t.TempDir()
	here := t.TempDir()
	t.Chdir(here)

	got := SourceKey(filepath.Join(outside, "ci.hcl"))

	if strings.HasPrefix(got, "..") {
		t.Fatalf("want an absolute path for a file outside the working directory, got %q", got)
	}

	if !filepath.IsAbs(filepath.FromSlash(got)) {
		t.Fatalf("want an absolute path, got %q", got)
	}
}

// TestReadSourceSetSpellsWhatSourceKeySpells pins the two against each other:
// the set is built from the parser's keys and compared against what a marker
// recorded, so they have to be produced the same way.
func TestReadSourceSetSpellsWhatSourceKeySpells(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	set := ReadSourceSet(map[string][]byte{
		"./src/ci.hcl":              []byte("workflow {}"),
		filepath.Join(dir, "a.hcl"): []byte("workflow {}"),
	})

	for _, want := range []string{"src/ci.hcl", "a.hcl"} {
		if _, ok := set[want]; !ok {
			t.Errorf("want %q in the read set, got %v", want, set)
		}
	}
}
