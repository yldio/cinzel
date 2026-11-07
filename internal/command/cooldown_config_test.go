// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/pin"
	githubprovider "github.com/yldio/cinzel/provider/github"
)

// readCooldown writes a config file, if any, and reads the policy the way
// upgrade does.
func readCooldown(t *testing.T, config string) (int, error) {
	t.Helper()

	var days int

	var err error

	withTempWorkingDir(t, func() {
		if config != "" {
			if writeErr := os.WriteFile(configFilename, []byte(config), 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
		}

		days, err = configuredCooldownDays()
	})

	return days, err
}

// The policy lives in the file rather than in the flag, because a cooldown is
// something a repository keeps: a scheduled upgrade job that has to remember
// --cooldown-days is one edit away from silently not having it.
func TestCooldownIsReadFromTheConfigFile(t *testing.T) {
	days, err := readCooldown(t, "github:\n  upgrade:\n    cooldown-days: 14\n")
	if err != nil {
		t.Fatal(err)
	}

	if days != 14 {
		t.Errorf("cooldown-days = %d, want 14", days)
	}
}

// Every path that means "no policy" has to reach 0 rather than an error, since
// most repositories have no config file at all and upgrade still has to run.
func TestNoCooldownConfigured(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
	}{
		{"no file", ""},
		{"empty file", "\n"},
		{"no github section", "gitlab:\n  parse:\n    directory: ci\n"},
		{"no upgrade section", "github:\n  parse:\n    directory: cinzel\n"},
		{"no cooldown key", "github:\n  upgrade:\n    something: else\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			days, err := readCooldown(t, tc.config)
			if err != nil {
				t.Fatalf("%s should not be an error: %v", tc.name, err)
			}

			if days != 0 {
				t.Errorf("cooldown-days = %d, want 0", days)
			}
		})
	}
}

// A quoted number is a string to YAML. Taking it would make the one spelling
// that looks identical to a reader behave differently, so it is refused by
// name rather than read as 0, which would be a policy silently not applied.
func TestACooldownThatIsNotAWholeNumberIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
	}{
		{"quoted", "github:\n  upgrade:\n    cooldown-days: \"14\"\n"},
		{"fractional", "github:\n  upgrade:\n    cooldown-days: 1.5\n"},
		{"a word", "github:\n  upgrade:\n    cooldown-days: fortnight\n"},
		{"a mapping", "github:\n  upgrade:\n    cooldown-days:\n      days: 14\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := readCooldown(t, tc.config)
			if err == nil {
				t.Fatalf("%s was accepted", tc.name)
			}

			if !strings.Contains(err.Error(), "cooldown-days") {
				t.Errorf("the error does not name the key: %v", err)
			}
		})
	}
}

// The flag is for a one-off, so it wins over the file when it is set, the way
// every other setting here resolves.
func TestTheFlagOverridesTheConfiguredCooldown(t *testing.T) {
	withTempWorkingDir(t, func() {
		if err := os.WriteFile(configFilename, []byte("github:\n  upgrade:\n    cooldown-days: 14\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		// --cooldown-days 0 is the way to turn the policy off for one run, so
		// it has to be distinguishable from the flag being absent.
		for _, tc := range []struct {
			args []string
			want time.Duration
		}{
			{[]string{"cinzel", "github", "upgrade", "--cooldown-days", "30"}, 30 * 24 * time.Hour},
			{[]string{"cinzel", "github", "upgrade", "--cooldown-days", "0"}, 0},
			{[]string{"cinzel", "github", "upgrade"}, 14 * 24 * time.Hour},
		} {
			got, err := cooldownFromArgs(t, tc.args)
			if err != nil {
				t.Fatal(err)
			}

			if got.For != tc.want {
				t.Errorf("%v gave a cooldown of %s, want %s", tc.args[3:], got.For, tc.want)
			}
		}
	})
}

// A negative window is not "no cooldown" — 0 already says that. It is a number
// whose author meant something the policy cannot carry out, so it is refused
// rather than clamped.
func TestANegativeCooldownIsRefused(t *testing.T) {
	withTempWorkingDir(t, func() {
		_, err := cooldownFromArgs(t, []string{"cinzel", "github", "upgrade", "--cooldown-days", "-5"})
		if !errors.Is(err, errNegativeCooldown) {
			t.Errorf("a negative cooldown = %v, want errNegativeCooldown", err)
		}
	})
}

// cooldownFromArgs parses a command line the way the CLI does and reads the
// policy out of it, so what is under test is the flag as urfave sees it
// rather than a hand-built Command.
func cooldownFromArgs(t *testing.T, args []string) (pin.Cooldown, error) {
	t.Helper()

	var (
		cooldown pin.Cooldown
		readErr  error
		ran      bool
	)

	app := NewWithErrWriter(new(bytes.Buffer), new(bytes.Buffer), "v.test")
	upgrade := app.upgradeCommand(githubprovider.New())

	upgrade.Action = func(_ context.Context, c *cli.Command) error {
		ran = true
		cooldown, readErr = upgradeCooldown(c)

		return nil
	}

	root := &cli.Command{
		Name:      "cinzel",
		Writer:    new(bytes.Buffer),
		ErrWriter: new(bytes.Buffer),
		Commands: []*cli.Command{
			{Name: "github", Commands: []*cli.Command{upgrade}},
		},
	}

	if err := root.Run(context.Background(), args); err != nil {
		t.Fatalf("parsing %v: %v", args, err)
	}

	if !ran {
		t.Fatalf("the upgrade action never ran for %v", args)
	}

	return cooldown, readErr
}

// A held release is the policy working, not a problem to go red over. If
// --check failed on one, a repository with a cooldown would have a permanently
// red check for the length of every window, and the way to green would be to
// abandon the policy.
//
// The distinction is only visible in the summary: held results are counted
// apart from upgraded ones, and only the upgraded count decides --check.
func TestCheckPassesOnAReleaseHeldInCooldown(t *testing.T) {
	out := new(bytes.Buffer)
	app := NewWithErrWriter(out, new(bytes.Buffer), "v.test")

	held := []pin.UpgradeResult{
		{Action: "actions/checkout", OldVersion: "v4", NewTag: "v5", HeldFor: 11 * 24 * time.Hour},
		{Action: "actions/setup-go", OldVersion: "v5", WasCurrent: true},
	}

	if err := app.printUpgradeSummary(held, true); err != nil {
		t.Errorf("upgrade --check over a held release = %v, want nil", err)
	}

	if !strings.Contains(out.String(), "1 held in cooldown") {
		t.Errorf("the summary does not count what was held: %q", out.String())
	}

	// An upgrade that is past its cooldown still fails the check, or the
	// above would be passing because nothing fails.
	outdated := []pin.UpgradeResult{
		{Action: "actions/checkout", OldVersion: "v4", NewTag: "v5", NewSHA: "abcdef1234567890abcdef1234567890abcdef12"},
	}

	if err := app.printUpgradeSummary(outdated, true); !errors.Is(err, errActionsOutdated) {
		t.Errorf("upgrade --check over an available upgrade = %v, want errActionsOutdated", err)
	}
}

// A run with no policy has no cooldown to hold anything in, so the summary
// does not carry a count of it.
func TestTheSummaryIsSilentAboutCooldownWhenNothingIsHeld(t *testing.T) {
	out := new(bytes.Buffer)
	app := NewWithErrWriter(out, new(bytes.Buffer), "v.test")

	if err := app.printUpgradeSummary([]pin.UpgradeResult{{Action: "actions/checkout", WasCurrent: true}}, false); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(out.String(), "cooldown") {
		t.Errorf("the summary mentions a cooldown that held nothing: %q", out.String())
	}
}
