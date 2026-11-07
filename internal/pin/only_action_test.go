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

const twoActions = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "setup" {
  uses {
    action  = "actions/setup-go"
    version = "v4"
  }
}
`

func upgradeOnly(t *testing.T, only string) ([]UpgradeResult, string, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "steps.hcl")

	if err := os.WriteFile(path, []byte(twoActions), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &stubUpgrader{
		latestTags: map[string]string{
			"actions/checkout": "v5",
			"actions/setup-go": "v6",
		},
		shas: map[string]string{
			"actions/checkout@v5": "1111111111111111111111111111111111111111",
			"actions/setup-go@v6": "2222222222222222222222222222222222222222",
		},
	}

	var buf bytes.Buffer

	results, err := UpgradeFile(context.Background(), path, resolver, &buf, UpgradeOptions{Only: only})
	if err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return results, buf.String(), string(after)
}

// Upgrading one action at a time is how a risky bump gets taken on its own,
// separately from the rest of the file. Without this the only way to move one
// action was to run the whole directory and revert what you did not want.
func TestOnlyActionUpgradesTheOneNamed(t *testing.T) {
	results, out, after := upgradeOnly(t, "actions/setup-go")

	if len(results) != 2 {
		t.Fatalf("expected both actions reported, got %d", len(results))
	}

	// Every action is still in the results. Dropping the ones not targeted
	// would make a targeted run and a one-action file look the same.
	var skipped, upgraded int

	for _, r := range results {
		if r.Skipped {
			skipped++

			continue
		}

		upgraded++

		if r.Action != "actions/setup-go" {
			t.Errorf("the run looked at %s, which it was not pointed at", r.Action)
		}
	}

	if skipped != 1 || upgraded != 1 {
		t.Errorf("skipped = %d, upgraded = %d, want 1 and 1", skipped, upgraded)
	}

	if !strings.Contains(after, "2222222222222222222222222222222222222222") {
		t.Errorf("the targeted action was not upgraded:\n%s", after)
	}

	// The other one is untouched, which is the whole point of the flag.
	if !strings.Contains(after, `action  = "actions/checkout"`) || strings.Contains(after, "1111111111111111111111111111111111111111") {
		t.Errorf("an action that was not targeted changed:\n%s", after)
	}

	if strings.Contains(out, "actions/checkout") {
		t.Errorf("the run reported an action it did not look at: %q", out)
	}
}

// The control: with no target, both move as before.
func TestNoOnlyActionUpgradesEverything(t *testing.T) {
	results, _, after := upgradeOnly(t, "")

	for _, r := range results {
		if r.Skipped {
			t.Errorf("%s was skipped with no target set", r.Action)
		}
	}

	if !strings.Contains(after, "1111111111111111111111111111111111111111") ||
		!strings.Contains(after, "2222222222222222222222222222222222222222") {
		t.Errorf("an untargeted run did not upgrade both:\n%s", after)
	}
}

// A name matching nothing has to be distinguishable from a file with nothing
// to do. Here that is every result coming back skipped; the command turns
// that into a failure.
func TestOnlyActionMatchingNothingSkipsEverything(t *testing.T) {
	results, _, after := upgradeOnly(t, "actions/chekout")

	for _, r := range results {
		if !r.Skipped {
			t.Errorf("%s was looked at for a name that matches nothing", r.Action)
		}
	}

	if after != twoActions {
		t.Errorf("a run that matched nothing wrote to the file:\n%s", after)
	}
}
