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
	"time"
)

const cooldownHCL = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

// at builds a Cooldown whose clock is fixed, so a test does not depend on how
// long it takes to run.
func at(window time.Duration, now time.Time) Cooldown {
	return Cooldown{For: window, Now: func() time.Time { return now }}
}

// upgradeWithCooldown runs one upgrade over a file holding actions/checkout@v4
// and reports what came back, what was printed, and what the file says after.
func upgradeWithCooldown(t *testing.T, published time.Time, cooldown Cooldown) ([]UpgradeResult, string, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "steps.hcl")

	if err := os.WriteFile(path, []byte(cooldownHCL), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &stubUpgrader{
		latestTags: map[string]string{"actions/checkout": "v5"},
		shas:       map[string]string{"actions/checkout@v5": "1111111111111111111111111111111111111111"},
		published:  map[string]time.Time{"actions/checkout": published},
	}

	var buf bytes.Buffer

	results, err := UpgradeFile(context.Background(), path, resolver, &buf, UpgradeOptions{Cooldown: cooldown})
	if err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return results, buf.String(), string(after)
}

// An upgrade takes whatever the newest tag is, so a release published minutes
// ago reaches CI before anyone has used it, and a tag that is later moved or
// pulled is moved inside that same window. The cooldown is the wait.
func TestAReleaseInsideTheCooldownIsNotTaken(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	results, out, after := upgradeWithCooldown(t, now.AddDate(0, 0, -3), at(14*24*time.Hour, now))

	if len(results) != 1 {
		t.Fatalf("expected one action, got %d", len(results))
	}

	if results[0].HeldFor == 0 {
		t.Error("a release 3 days old under a 14 day cooldown was taken")
	}

	// The SHA is reported, since the run resolved it on the way to deciding.
	// What makes this a hold is the file, not the absence of a value.
	if results[0].NewSHA == "" {
		t.Error("a held action does not report the commit it would move to")
	}

	if after != cooldownHCL {
		t.Errorf("a held action was written to the file:\n%s", after)
	}

	if strings.Contains(after, results[0].NewSHA) {
		t.Errorf("the held SHA reached the file:\n%s", after)
	}

	// The reason has to be on screen. A run that silently declines an upgrade
	// is indistinguishable from one that found none, and the next person to
	// look wonders why the bump never happens.
	if !strings.Contains(out, "held actions/checkout") {
		t.Errorf("the run did not say it held anything: %q", out)
	}

	// What the wait ends in, named. "published 65d ago, 24d left" made the
	// reader work out both the window and the date it clears, and the two
	// floored figures did not add up to the window either.
	for _, want := range []string{
		"published 2026-11-03",
		"inside the 14d cooldown",
		"upgrades on or after 2026-11-17",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the held line does not say %q:\n%s", want, out)
		}
	}
}

// The control. The same release, past the window, upgrades as it always did.
func TestAReleasePastTheCooldownIsTaken(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	results, out, after := upgradeWithCooldown(t, now.AddDate(0, 0, -30), at(14*24*time.Hour, now))

	if results[0].HeldFor != 0 {
		t.Errorf("a release 30 days old under a 14 day cooldown was held for %s", results[0].HeldFor)
	}

	if results[0].NewTag != "v5" {
		t.Errorf("expected the upgrade to v5, got %q", results[0].NewTag)
	}

	if !strings.Contains(after, "1111111111111111111111111111111111111111") {
		t.Errorf("the upgrade was not written:\n%s", after)
	}

	if !strings.Contains(out, "upgraded actions/checkout") {
		t.Errorf("expected an upgrade line, got %q", out)
	}
}

// The zero Cooldown is what every caller that has no policy passes, and it has
// to behave exactly as upgrade did before any of this existed.
func TestNoCooldownTakesEverything(t *testing.T) {
	results, _, after := upgradeWithCooldown(t, time.Now().Add(-time.Minute), Cooldown{})

	if results[0].HeldFor != 0 {
		t.Error("a release published a minute ago was held with no policy set")
	}

	if !strings.Contains(after, "1111111111111111111111111111111111111111") {
		t.Error("the upgrade was not written with no policy set")
	}
}

// GitHub leaves published_at null on a draft release. An unknown date is not
// evidence that something is new, and holding on one would stall an upgrade
// for a reason the run cannot put on screen.
func TestAReleaseWithNoDateIsTaken(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	results, _, after := upgradeWithCooldown(t, time.Time{}, at(14*24*time.Hour, now))

	if results[0].HeldFor != 0 {
		t.Error("a release with no publication date was held")
	}

	if !strings.Contains(after, "1111111111111111111111111111111111111111") {
		t.Error("a release with no publication date was not written")
	}
}

// An action already on the newest tag is current, not held. Checking the
// cooldown first would report it as held for as long as the window lasted,
// which reads as an upgrade waiting that is not.
func TestAnActionAlreadyOnTheLatestIsCurrentNotHeld(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "steps.hcl")

	content := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v5"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &stubUpgrader{
		latestTags: map[string]string{"actions/checkout": "v5"},
		shas:       map[string]string{"actions/checkout@v5": "1111111111111111111111111111111111111111"},
		published:  map[string]time.Time{"actions/checkout": now.AddDate(0, 0, -1)},
	}

	var buf bytes.Buffer

	results, err := UpgradeFile(context.Background(), path, resolver, &buf, UpgradeOptions{Cooldown: at(14*24*time.Hour, now)})
	if err != nil {
		t.Fatal(err)
	}

	if !results[0].WasCurrent {
		t.Errorf("an action on the latest tag was reported as %+v", results[0])
	}

	if results[0].HeldFor != 0 {
		t.Error("an action on the latest tag was reported as held")
	}
}

func TestRoundDuration(t *testing.T) {
	for _, tc := range []struct {
		in   time.Duration
		want string
	}{
		{14 * 24 * time.Hour, "14d"},
		{47 * time.Hour, "47h0m0s"},
		{90 * time.Minute, "2h0m0s"},
		{30 * time.Second, "1m0s"},
	} {
		if got := roundDuration(tc.in); got != tc.want {
			t.Errorf("roundDuration(%s) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
