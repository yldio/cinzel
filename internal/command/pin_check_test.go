// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/internal/pin"
	"github.com/yldio/cinzel/provider"
	githubprovider "github.com/yldio/cinzel/provider/github"
)

// runCommandLineCapturingOutput is runCommandLine with the writer kept, for
// the one assertion that is about what the summary says rather than what the
// process exits with.
func runCommandLineCapturingOutput(t *testing.T, args []string) (string, error) {
	t.Helper()

	out := new(bytes.Buffer)
	app := NewWithErrWriter(out, new(bytes.Buffer), "v.test")

	err := app.Execute(args, []provider.Provider{githubprovider.New()})

	return out.String(), err
}

// --dry-run says what a run would do and exits 0 either way, so a CI step
// cannot gate on it: a workflow holding an action on a moving tag passed a
// "cinzel github pin --dry-run" step exactly as a fully pinned one did. The
// resolutions were printed, and the exit code said nothing about them.
//
// --check is the question rather than the report: an action that had to be
// resolved was not pinned when the run started, and the run wrote nothing, so
// it still is not. That is a failure, the way "fmt --check" fails on a file it
// would rewrite and "diff" fails on YAML that drifted.
//
// Resolving reaches the GitHub API, so the cases here are the ones that answer
// offline. An action on a full commit SHA is already pinned and skipped
// without a request; a local action names no repository and fails without one.
func TestPinCheckExitCodes(t *testing.T) {
	t.Run("a pinned tree returns nil", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			writeActionDir(t, pinnedAction)

			if err := runCommandLine(t, []string{
				"cinzel", "github", "pin", "--directory", "actions", "--check",
			}); err != nil {
				t.Errorf("pin --check over a pinned tree = %v, want nil", err)
			}
		})
	})

	t.Run("an action that could not be resolved outranks the check", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			writeActionDir(t, unresolvableAction)

			err := runCommandLine(t, []string{
				"cinzel", "github", "pin", "--directory", "actions", "--check",
			})

			// A run that could not reach an action does not know whether it
			// is pinned, so the failure is the more honest answer of the two.
			if !errors.Is(err, errPinFailed) {
				t.Errorf("pin --check over an unreachable action = %v, want errPinFailed", err)
			}
		})
	})
}

// Same shape for upgrade. Its offline cases are thinner: it asks for the
// latest release whatever the version says, so even a SHA-pinned action goes
// to the network. A step with no uses block has nothing to ask about.
func TestUpgradeCheckExitCodes(t *testing.T) {
	t.Run("a tree with nothing to upgrade returns nil", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			writeActionDir(t, noActionAtAll)

			if err := runCommandLine(t, []string{
				"cinzel", "github", "upgrade", "--directory", "actions", "--check",
			}); err != nil {
				t.Errorf("upgrade --check over a current tree = %v, want nil", err)
			}
		})
	})

	t.Run("an action that could not be resolved fails", func(t *testing.T) {
		withTempWorkingDir(t, func() {
			writeActionDir(t, unresolvableAction)

			err := runCommandLine(t, []string{
				"cinzel", "github", "upgrade", "--directory", "actions", "--check",
			})

			if !errors.Is(err, errUpgradeFailed) {
				t.Errorf("upgrade --check over an unreachable action = %v, want errUpgradeFailed", err)
			}
		})
	})
}

// An action on a tag, for the one case below that has to resolve something.
const taggedAction = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

// --check implies --dry-run, and a check that writes what it was asked to
// look for passes on its own side effect: the second run of a CI step would
// find the file pinned because the first run pinned it.
//
// A run that writes nothing is not on its own evidence of that — a run that
// resolved nothing leaves the file alone too. So this one has to resolve, and
// resolving is a request to the GitHub API. Pin reads a 24h cache first, so
// seeding it is the offline way in. The assertion on the summary is what
// guards the seed: if the key ever stops matching, the run reports nothing
// pinned and the test says so rather than passing on a resolution it never
// made.
//
// Upgrade has no equivalent. It asks for the latest release through an
// uncached resolver, so there is no offline path on which it rewrites a file.
func TestPinCheckWritesNothingToAFileItWouldPin(t *testing.T) {
	withTempWorkingDir(t, func() {
		seedPinCache(t, "actions", "checkout", "v4", "abcdef1234567890abcdef1234567890abcdef12")
		writeActionDir(t, taggedAction)

		out, err := runCommandLineCapturingOutput(t, []string{
			"cinzel", "github", "pin", "--directory", "actions", "--check",
		})

		if !errors.Is(err, errActionsNotPinned) {
			t.Fatalf("pin --check over an action on a tag = %v, want errActionsNotPinned", err)
		}

		if !strings.Contains(out, "actions/checkout@v4") {
			t.Fatalf("the run resolved nothing, so it proves nothing about writing:\n%s", out)
		}

		if after := readFile(t, "actions/steps.hcl"); after != taggedAction {
			t.Errorf("--check wrote the SHA it resolved:\n%s", after)
		}
	})
}

// seedPinCache writes the entry pin's cached resolver reads for one tag, so
// the command resolves it without a request. The key mirrors the one in
// internal/pin; the caller asserts the seed took.
func seedPinCache(t *testing.T, owner, repo, tag, sha string) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", "")

	base, err := os.UserCacheDir()
	if err != nil {
		t.Skipf("no user cache directory to seed: %v", err)
	}

	dir := filepath.Join(base, "cinzel", "pins")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	key := sha256.Sum256([]byte(fmt.Sprintf("%s/%s@%s", owner, repo, tag)))

	if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%x", key[:16])), []byte(sha), 0o600); err != nil {
		t.Fatal(err)
	}
}

// The summary is read by whoever is looking at the failed step, so under
// check it has to describe what the run found rather than what it did. "2
// pinned" over a file the run deliberately left alone is the opposite of the
// finding.
func TestCheckSummaryDoesNotClaimAWriteItDidNotMake(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeActionDir(t, pinnedAction)

		out, err := runCommandLineCapturingOutput(t, []string{
			"cinzel", "github", "pin", "--directory", "actions", "--check",
		})
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(out, "0 not pinned") {
			t.Errorf("check summary = %q, want it to count what is not pinned", out)
		}
	})
}

// The case the CLI tests above cannot reach. Resolving a tag is a request to
// the GitHub API, so a run that actually pins something needs the network,
// and the whole point of check is what the command does with that result.
//
// The summary is where the decision is made, so that is what this drives. A
// PinResult carrying a SHA and no error is an action that was on a tag and had
// to be resolved: under check it is the finding, and the run has to end
// non-zero over it.
func TestPinSummaryFailsUnderCheckOnAnActionItResolved(t *testing.T) {
	app := NewWithErrWriter(new(bytes.Buffer), new(bytes.Buffer), "v.test")

	resolved := []pin.PinResult{
		{Action: "actions/checkout", Tag: "v4", SHA: "abcdef1234567890abcdef1234567890abcdef12"},
		{Action: "actions/setup-go", SHA: "abcdef1234567890abcdef1234567890abcdef12", WasAlready: true},
	}

	if err := app.printPinSummary(resolved, false); err != nil {
		t.Errorf("without check, a resolved action is the work: got %v, want nil", err)
	}

	err := app.printPinSummary(resolved, true)
	if !errors.Is(err, errActionsNotPinned) {
		t.Errorf("pin --check over a resolved action = %v, want errActionsNotPinned", err)
	}
}

// Same for upgrade: a result with a NewTag and no error is an action that had
// a newer release than the file held.
func TestUpgradeSummaryFailsUnderCheckOnAnActionItUpgraded(t *testing.T) {
	app := NewWithErrWriter(new(bytes.Buffer), new(bytes.Buffer), "v.test")

	outdated := []pin.UpgradeResult{
		{Action: "actions/checkout", OldVersion: "v4", NewTag: "v5", NewSHA: "abcdef1234567890abcdef1234567890abcdef12"},
		{Action: "actions/setup-go", OldVersion: "v5", WasCurrent: true},
	}

	if err := app.printUpgradeSummary(outdated, false); err != nil {
		t.Errorf("without check, an upgraded action is the work: got %v, want nil", err)
	}

	err := app.printUpgradeSummary(outdated, true)
	if !errors.Is(err, errActionsOutdated) {
		t.Errorf("upgrade --check over an outdated action = %v, want errActionsOutdated", err)
	}
}
