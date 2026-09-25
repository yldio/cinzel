// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// A file neither reader could open was reported as a warning and dropped, and
// both summaries count results, so a directory holding one unparseable file and
// one already-pinned action read "0 pinned, 1 already pinned, 0 failed" at exit
// 0. Nothing in that says a file was skipped.
func TestAFileThatCannotBeReadIsCountedAsAFailure(t *testing.T) {
	const pinned = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "abcdef1234567890abcdef1234567890abcdef12"
  }
}
`

	const broken = `step "oops" {
  uses {
    action = "actions/checkout"
`

	t.Run("pin", func(t *testing.T) {
		dir := writeTwoFiles(t, pinned, broken)

		var buf bytes.Buffer

		results, err := PinDirectory(context.Background(), dir, &mockResolver{}, &buf, true)
		if err != nil {
			t.Fatal(err)
		}

		want := filepath.Join(dir, "b.hcl")

		if !hasFailureFor(want, len(results), func(i int) (string, error) {
			return results[i].Action, results[i].Error
		}) {
			t.Errorf("no failed result for %s, got %+v", want, results)
		}
	})

	t.Run("upgrade", func(t *testing.T) {
		dir := writeTwoFiles(t, pinned, broken)

		var buf bytes.Buffer

		results, err := UpgradeDirectory(context.Background(), dir, &stubUpgrader{}, &buf, true)
		if err != nil {
			t.Fatal(err)
		}

		want := filepath.Join(dir, "b.hcl")

		if !hasFailureFor(want, len(results), func(i int) (string, error) {
			return results[i].Action, results[i].Error
		}) {
			t.Errorf("no failed result for %s, got %+v", want, results)
		}
	})
}

func writeTwoFiles(t *testing.T, good, broken string) string {
	t.Helper()

	dir := t.TempDir()

	for name, content := range map[string]string{"a.hcl": good, "b.hcl": broken} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

// hasFailureFor reports whether the results name the given path as having
// failed, which is what both summaries count.
func hasFailureFor(path string, n int, at func(int) (string, error)) bool {
	for i := range n {
		if action, err := at(i); action == path && err != nil {
			return true
		}
	}

	return false
}
