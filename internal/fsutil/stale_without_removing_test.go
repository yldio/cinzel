// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

// StaleGeneratedYAML answers what the prune would remove without removing it,
// which is what lets "cinzel <provider> diff" report an orphan left by a
// rename. The prune is this walk plus a remove, so the two have to agree on
// every file: anything named here and not deleted there, or the other way
// round, is a diff that reports something a parse would not do.
func TestStaleGeneratedYAMLNamesWhatThePruneRemoves(t *testing.T) {
	generated := func(name string) string {
		return "# generated-by: cinzel\n# cinzel-provider: github\nname: " + name + "\n"
	}

	setUp := func(t *testing.T) (dir string, current string, stale string, manual string) {
		t.Helper()

		dir = t.TempDir()
		current = filepath.Join(dir, "current.yaml")
		stale = filepath.Join(dir, "stale.yaml")
		manual = filepath.Join(dir, "manual.yaml")

		for path, content := range map[string]string{
			current: generated("current"),
			stale:   generated("stale"),
			manual:  "name: manual\n",
		} {
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		return dir, current, stale, manual
	}

	t.Run("names the stale file and leaves every file alone", func(t *testing.T) {
		dir, current, stale, manual := setUp(t)

		got, err := StaleGeneratedYAML(dir, map[string]struct{}{filepath.Clean(current): {}}, "github", nil)
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 1 {
			t.Fatalf("StaleGeneratedYAML() = %v, want just the stale file", got)
		}

		if filepath.Base(got[0]) != "stale.yaml" {
			t.Errorf("named %s, want stale.yaml", got[0])
		}

		for _, path := range []string{current, stale, manual} {
			if _, err := os.Stat(path); err != nil {
				t.Errorf("%s was removed by a call that only reports: %v", filepath.Base(path), err)
			}
		}
	})

	t.Run("agrees with what the prune deletes", func(t *testing.T) {
		dir, current, _, _ := setUp(t)

		outputs := map[string]struct{}{filepath.Clean(current): {}}

		named, err := StaleGeneratedYAML(dir, outputs, "github", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := PruneStaleGeneratedYAML(dir, outputs, "github", nil); err != nil {
			t.Fatal(err)
		}

		for _, path := range named {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("%s was named stale and is still there", path)
			}
		}

		left, err := StaleGeneratedYAML(dir, outputs, "github", nil)
		if err != nil {
			t.Fatal(err)
		}

		if len(left) != 0 {
			t.Errorf("the prune left %v behind", left)
		}
	})

	// A planned output that is not on disk is what diff calls "would create".
	// sameAsCurrentOutput stats every candidate, and under diff nothing has
	// been written, so a failed stat has to read as "not this file" rather
	// than taking the whole walk down.
	t.Run("a current output that does not exist yet", func(t *testing.T) {
		dir, _, stale, _ := setUp(t)

		unwritten := filepath.Join(dir, "not-written-yet.yaml")

		got, err := StaleGeneratedYAML(dir, map[string]struct{}{filepath.Clean(unwritten): {}}, "github", nil)
		if err != nil {
			t.Fatalf("StaleGeneratedYAML() = %v, want a path that does not exist to be skipped", err)
		}

		if len(got) != 2 {
			t.Fatalf("StaleGeneratedYAML() = %v, want both generated files named stale", got)
		}

		if _, err := os.Stat(stale); err != nil {
			t.Errorf("a reporting call removed %s: %v", stale, err)
		}
	})

	t.Run("a missing output directory names nothing", func(t *testing.T) {
		got, err := StaleGeneratedYAML(filepath.Join(t.TempDir(), "absent"), map[string]struct{}{}, "github", nil)
		if err != nil {
			t.Fatalf("StaleGeneratedYAML() = %v, want a missing directory to be no error", err)
		}

		if len(got) != 0 {
			t.Errorf("StaleGeneratedYAML() = %v, want nothing", got)
		}
	})
}
