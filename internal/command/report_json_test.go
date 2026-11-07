// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yldio/cinzel/internal/pin"
	"github.com/yldio/cinzel/provider"
	githubprovider "github.com/yldio/cinzel/provider/github"
)

// providersForTest is the provider set the real binary registers.
func providersForTest() []provider.Provider {
	return []provider.Provider{githubprovider.New()}
}

// decodeReport runs one report and parses what landed on stdout, which is the
// whole contract: a consumer pipes this into a parser, so anything that is not
// a document is a failure however readable it is.
func decodeReport(t *testing.T, out *bytes.Buffer) map[string]any {
	t.Helper()

	var report map[string]any

	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out.String())
	}

	return report
}

func TestPinJSONReport(t *testing.T) {
	out := new(bytes.Buffer)
	app := NewWithErrWriter(out, new(bytes.Buffer), "v.test")

	results := []pin.PinResult{
		{Action: "actions/checkout", Tag: "v4", SHA: "abcdef1234567890abcdef1234567890abcdef12"},
		{Action: "actions/setup-go", SHA: "1111111111111111111111111111111111111111", WasAlready: true},
		{Action: "./local/build", Tag: "v1", Error: errors.New("names no repository")},
	}

	err := app.printPinJSON(results, true, false)

	// A failed action is still a failed run, whatever the output format.
	if !errors.Is(err, errPinFailed) {
		t.Errorf("a report with a failure returned %v, want errPinFailed", err)
	}

	report := decodeReport(t, out)

	if report["command"] != "pin" || report["dryRun"] != true {
		t.Errorf("envelope = %v", report)
	}

	summary, ok := report["summary"].(map[string]any)
	if !ok {
		t.Fatalf("no summary in %v", report)
	}

	for key, want := range map[string]float64{"pinned": 1, "alreadyPinned": 1, "failed": 1} {
		if summary[key] != want {
			t.Errorf("summary[%q] = %v, want %v", key, summary[key], want)
		}
	}

	actions, ok := report["actions"].([]any)
	if !ok || len(actions) != 3 {
		t.Fatalf("expected three actions, got %v", report["actions"])
	}

	// The reason a failure happened belongs in the document. A consumer that
	// captured stdout should not have to parse the human output as well to
	// find out why something could not be pinned.
	failed, ok := actions[2].(map[string]any)
	if !ok {
		t.Fatalf("third action is not an object: %v", actions[2])
	}

	if failed["status"] != "failed" || failed["error"] != "names no repository" {
		t.Errorf("the failure does not carry its reason: %v", failed)
	}
}

func TestUpgradeJSONReport(t *testing.T) {
	out := new(bytes.Buffer)
	app := NewWithErrWriter(out, new(bytes.Buffer), "v.test")

	published := time.Date(2026, 9, 1, 21, 31, 19, 0, time.UTC)

	results := []pin.UpgradeResult{
		{Action: "actions/checkout", OldVersion: "v4", NewTag: "v5", NewSHA: "abcdef1234567890abcdef1234567890abcdef12"},
		{Action: "actions/setup-go", OldVersion: "v5", WasCurrent: true},
		{
			Action:     "actions/deploy-pages",
			OldVersion: "d6db90164ac5ed86f2b6aed7e0febac5b3c0c03e",
			OldTag:     "v4",
			NewTag:     "v5.0.1",
			HeldFor:    24 * 24 * time.Hour,
			Published:  published,
			EligibleAt: published.AddDate(0, 0, 90),
		},
	}

	if err := app.printUpgradeJSON(results, true, false); err != nil {
		t.Fatalf("a report with no failure returned %v", err)
	}

	report := decodeReport(t, out)

	summary, ok := report["summary"].(map[string]any)
	if !ok {
		t.Fatalf("no summary in %v", report)
	}

	for key, want := range map[string]float64{"upgraded": 1, "alreadyCurrent": 1, "failed": 0, "held": 1} {
		if summary[key] != want {
			t.Errorf("summary[%q] = %v, want %v", key, summary[key], want)
		}
	}

	actions, _ := report["actions"].([]any)

	held, ok := actions[2].(map[string]any)
	if !ok {
		t.Fatalf("third action is not an object: %v", actions[2])
	}

	// A date, not a remaining duration. A duration is read against the clock
	// at the moment of the run, so a report written now and read tomorrow
	// says something that is no longer true.
	if held["status"] != "held" || held["eligibleAt"] != "2026-11-30" || held["published"] != "2026-09-01" {
		t.Errorf("the held action does not carry its dates: %v", held)
	}

	// Each end of the move is a release and a commit. "from" used to carry
	// whatever the line held, so it was a tag on an unpinned action and a
	// SHA on a pinned one, and a consumer had to know which before it could
	// read it. A pinned line names its release only in the comment.
	if held["from"] != "v4" || held["fromSHA"] != "d6db90164ac5ed86f2b6aed7e0febac5b3c0c03e" {
		t.Errorf("the from end is not split into release and commit: %v", held)
	}

	upgraded, ok := actions[0].(map[string]any)
	if !ok {
		t.Fatalf("first action is not an object: %v", actions[0])
	}

	if upgraded["to"] != "v5" || upgraded["toSHA"] != "abcdef1234567890abcdef1234567890abcdef12" {
		t.Errorf("the to end is not split into release and commit: %v", upgraded)
	}

	// An unpinned action has no commit on the from end until one is resolved,
	// so the field is absent rather than empty.
	if _, present := upgraded["fromSHA"]; present {
		t.Errorf("an action on a tag reported a fromSHA: %v", upgraded)
	}

	if upgraded["from"] != "v4" {
		t.Errorf("from = %v, want the tag the line held", upgraded["from"])
	}
}

// The flag changes how a run is reported, not what it concludes. If the two
// paths could exit differently over the same results, a pipeline that moved to
// --json would quietly change its pass/fail behaviour.
//
// Both now call one outcome function, so this guards against them drifting
// apart rather than against that function being wrong — breaking it fails
// both sides here and neither is reported. What the outcome itself has to be
// is in pin_check_test.go, which drives the real CLI.
func TestJSONAndTextAgreeOnTheExitCode(t *testing.T) {
	pins := [][]pin.PinResult{
		{{Action: "a", SHA: "1111111111111111111111111111111111111111", WasAlready: true}},
		{{Action: "a", Tag: "v4", SHA: "1111111111111111111111111111111111111111"}},
		{{Action: "a", Tag: "v4", Error: errors.New("boom")}},
	}

	for _, check := range []bool{false, true} {
		for i, results := range pins {
			text := NewWithErrWriter(new(bytes.Buffer), new(bytes.Buffer), "v.test")
			asJSON := NewWithErrWriter(new(bytes.Buffer), new(bytes.Buffer), "v.test")

			textErr := text.printPinSummary(results, check)
			jsonErr := asJSON.printPinJSON(results, false, check)

			if (textErr == nil) != (jsonErr == nil) {
				t.Errorf("pin case %d, check=%v: text = %v, json = %v", i, check, textErr, jsonErr)
			}
		}
	}

	upgrades := [][]pin.UpgradeResult{
		{{Action: "a", WasCurrent: true}},
		{{Action: "a", NewTag: "v5", NewSHA: "1111111111111111111111111111111111111111"}},
		{{Action: "a", Error: errors.New("boom")}},
		{{Action: "a", NewTag: "v5", HeldFor: time.Hour}},
	}

	for _, check := range []bool{false, true} {
		for i, results := range upgrades {
			text := NewWithErrWriter(new(bytes.Buffer), new(bytes.Buffer), "v.test")
			asJSON := NewWithErrWriter(new(bytes.Buffer), new(bytes.Buffer), "v.test")

			textErr := text.printUpgradeSummary(results, check)
			jsonErr := asJSON.printUpgradeJSON(results, false, check)

			if (textErr == nil) != (jsonErr == nil) {
				t.Errorf("upgrade case %d, check=%v: text = %v, json = %v", i, check, textErr, jsonErr)
			}
		}
	}
}

// Stdout has to be a document and nothing else. The per-action lines share the
// writer with the report, so interleaved they would produce something no
// parser accepts — which is the one thing the flag exists to avoid.
func TestProgressStaysOutOfTheJSONOnStdout(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeActionDir(t, unresolvableAction)

		out := new(bytes.Buffer)
		errOut := new(bytes.Buffer)
		app := NewWithErrWriter(out, errOut, "v.test")

		// A local action fails without a request, so this needs no network.
		_ = app.Execute(
			[]string{"cinzel", "github", "pin", "--directory", "actions", "--check", "--json"},
			providersForTest(),
		)

		if json.Unmarshal(out.Bytes(), new(map[string]any)) != nil {
			t.Fatalf("stdout is not a JSON document:\n%s", out.String())
		}

		if strings.Contains(out.String(), "warning:") {
			t.Errorf("a progress line landed in the JSON:\n%s", out.String())
		}

		if !strings.Contains(errOut.String(), "warning: could not pin") {
			t.Errorf("the warning did not reach stderr:\n%s", errOut.String())
		}
	})
}

// A name that matches nothing must not report success. Every action comes
// back skipped, which is no upgrade and no failure and so reads as a clean
// run — the same trap "fmt" fell into when a wrong path reported a clean
// check of no files.
//
// Driven through the real CLI, so what it asserts on is the error the process
// exits with. A local action is used because it resolves offline: the run
// never reaches it, since the name does not match, and the test needs no
// network either way.
func TestOnlyActionMatchingNothingFails(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeActionDir(t, unresolvableAction)

		err := runCommandLine(t, []string{
			"cinzel", "github", "upgrade", "--directory", "actions",
			"--dry-run", "--only-action", "actions/typo-here",
		})

		if !errors.Is(err, errNoSuchAction) {
			t.Errorf("a target matching nothing = %v, want errNoSuchAction", err)
		}

		// The control: the name that is there is not an error, so the above
		// is not passing because every targeted run fails.
		err = runCommandLine(t, []string{
			"cinzel", "github", "upgrade", "--directory", "actions",
			"--dry-run", "--only-action", "./.github/actions/build",
		})

		if errors.Is(err, errNoSuchAction) {
			t.Errorf("a target that matches reported that it matched nothing: %v", err)
		}
	})
}

// Skipped actions stay in the report and are counted, so the numbers add up
// to what is in the file rather than quietly omitting the rest.
func TestSkippedActionsAreCountedAndReported(t *testing.T) {
	out := new(bytes.Buffer)
	app := NewWithErrWriter(out, new(bytes.Buffer), "v.test")

	results := []pin.UpgradeResult{
		{Action: "actions/checkout", OldVersion: "v4", Skipped: true},
		{Action: "actions/setup-go", OldVersion: "v4", NewTag: "v6", NewSHA: "abcdef1234567890abcdef1234567890abcdef12"},
	}

	if err := app.printUpgradeJSON(results, true, false); err != nil {
		t.Fatal(err)
	}

	report := decodeReport(t, out)

	summary, ok := report["summary"].(map[string]any)
	if !ok {
		t.Fatalf("no summary in %v", report)
	}

	if summary["skipped"] != float64(1) || summary["upgraded"] != float64(1) {
		t.Errorf("summary = %v, want 1 skipped and 1 upgraded", summary)
	}

	actions, _ := report["actions"].([]any)

	skipped, ok := actions[0].(map[string]any)
	if !ok {
		t.Fatalf("first action is not an object: %v", actions[0])
	}

	if skipped["status"] != "skipped" {
		t.Errorf("status = %v, want skipped", skipped["status"])
	}
}
