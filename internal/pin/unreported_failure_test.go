// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// A version or an action that cannot be pinned is counted in the summary but
// was never named, so a run ending "2 failed" left the user to find which two
// lines those were. A failed resolve has always printed a warning naming the
// action; these two took the same path through the results and printed
// nothing.
func TestAnActionThatCannotBePinnedIsNamed(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "actions/checkout"
    version = "main"
  }
}

step "b" {
  uses {
    action  = "./.github/actions/build"
    version = "v1"
  }
}
`)

	resolver := &mockResolver{shas: map[string]string{}}

	var buf bytes.Buffer

	results, err := PinFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	out := buf.String()

	for _, want := range []string{"actions/checkout", "main", "./.github/actions/build"} {
		if !strings.Contains(out, want) {
			t.Errorf("the output does not name %q:\n%s", want, out)
		}
	}
}

// The same on the upgrade side, where an action naming no repository is the
// only one of the two that can be reached: a version that is not a tag is
// still upgradable.
func TestAnActionThatCannotBeUpgradedIsNamed(t *testing.T) {
	path := writeHCL(t, `step "b" {
  uses {
    action  = "./.github/actions/build"
    version = "v1"
  }
}
`)

	resolver := &upgraderStub{}

	var buf bytes.Buffer

	if _, err := UpgradeFile(context.Background(), path, resolver, &buf, false); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "./.github/actions/build") {
		t.Errorf("the output does not name the action:\n%s", buf.String())
	}
}
