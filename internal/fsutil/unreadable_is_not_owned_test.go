// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// A file in the output directory that cannot be read is not cinzel's, so the
// prune passes over it. Reading it as an error instead ended the whole run: a
// parse that had already written its YAML exited 1 over a file the author put
// there, and the message sent them to the issue tracker.
//
// bufio.Scanner refuses a line over 64KB, which is the shape that reaches this
// through a plain readable file. Nothing cinzel writes can be that file — its
// markers are the first two lines, so the scan returns before any long line —
// which is exactly why the file in question is always someone else's.
func TestAFileThatCannotBeReadIsLeftAlone(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content []byte
		mode    os.FileMode
		skip    string
	}{
		{
			name:    "a line longer than the scanner accepts",
			content: []byte("# " + strings.Repeat("x", 100000) + "\nname: mine\n"),
			mode:    0600,
		},
		{
			name:    "unreadable permissions",
			content: []byte("name: mine\n"),
			mode:    0000,
			skip:    "windows",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip == runtime.GOOS {
				t.Skipf("file modes do not deny reads on %s", runtime.GOOS)
			}

			dir := t.TempDir()

			theirs := filepath.Join(dir, "theirs.yaml")
			if err := os.WriteFile(theirs, tc.content, 0600); err != nil {
				t.Fatal(err)
			}

			if err := os.Chmod(theirs, tc.mode); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = os.Chmod(theirs, 0600) })

			// One output from this run, and one of cinzel's own left behind by
			// an earlier one. The stale file has to still go: a prune that
			// stops deleting anything would pass the first assertion alone.
			current := filepath.Join(dir, "current.yaml")
			if err := os.WriteFile(current, PrependGeneratedMarker([]byte("name: current\n"), "github"), 0600); err != nil {
				t.Fatal(err)
			}

			stale := filepath.Join(dir, "stale.yaml")
			if err := os.WriteFile(stale, PrependGeneratedMarker([]byte("name: stale\n"), "github"), 0600); err != nil {
				t.Fatal(err)
			}

			if err := PruneStaleGeneratedYAML(dir, map[string]struct{}{current: {}}, "github"); err != nil {
				t.Fatalf("the prune failed over a file it does not own: %v", err)
			}

			if _, err := os.Stat(theirs); err != nil {
				t.Errorf("a file cinzel does not own was removed: %v", err)
			}

			if _, err := os.Stat(current); err != nil {
				t.Errorf("this run's own output was removed: %v", err)
			}

			if _, err := os.Stat(stale); !os.IsNotExist(err) {
				t.Errorf("the stale file was left behind, so nothing was pruned at all")
			}
		})
	}
}
