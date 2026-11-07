// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/pin"
	"github.com/yldio/cinzel/provider"
)

func (cmd *Cli) upgradeCommand(p provider.Provider) *cli.Command {
	return &cli.Command{
		Name:  "upgrade",
		Usage: "Upgrade GitHub Actions to their latest versions and pin to SHAs",
		Action: func(ctx context.Context, c *cli.Command) error {
			filePath := c.String("file")
			dirPath := c.String("directory")
			check := c.Bool("check")
			parse := c.Bool("parse")

			// See pin: a check that writes what it was asked to look for
			// passes on its own side effect.
			dryRun := c.Bool("dry-run") || check

			if filePath == "" && dirPath == "" {
				dirPath = "cinzel"
			}

			if filePath != "" {
				if err := validateRelativePath(filePath); err != nil {
					return fmt.Errorf("--file: %w", err)
				}

				if err := requireFileNotDirectory(filePath); err != nil {
					return err
				}
			}

			if dirPath != "" {
				if err := validateRelativePath(dirPath); err != nil {
					return fmt.Errorf("--directory: %w", err)
				}
			}

			cooldown, err := upgradeCooldown(c)
			if err != nil {
				return err
			}

			// No cache wrapper — upgrade checks latest releases which should not
			// be served from a 24h cache.
			resolver := pin.NewGitHubResolver("")

			var results []pin.UpgradeResult

			asJSON := c.Bool("json")

			// See pin: under --json the per-action lines go to stderr, or
			// stdout is not a document.
			progress := cmd.Writer
			if asJSON {
				progress = c.Root().ErrWriter
			}

			only := c.String("only-action")

			opts := pin.UpgradeOptions{DryRun: dryRun, Cooldown: cooldown, Only: only}

			if filePath != "" {
				results, err = pin.UpgradeFile(ctx, filePath, resolver, progress, opts)
			} else {
				results, err = pin.UpgradeDirectory(ctx, dirPath, resolver, progress, opts)
			}

			if err != nil {
				return err
			}

			// Before the report, since the report of a run that looked at
			// nothing is not worth reading.
			if only != "" && !anyConsidered(results) {
				return fmt.Errorf("%w: %s", errNoSuchAction, only)
			}

			var summaryErr error

			if asJSON {
				summaryErr = cmd.printUpgradeJSON(results, dryRun, check)
			} else {
				summaryErr = cmd.printUpgradeSummary(results, check)
			}

			if !parse || dryRun {
				return summaryErr
			}

			// Check if anything was actually upgraded.
			upgraded := false

			for _, r := range results {
				if r.Error == nil && !r.WasCurrent {
					upgraded = true

					break
				}
			}

			if !upgraded {
				return summaryErr
			}

			parseDir := dirPath
			if filePath != "" {
				// If a single file was upgraded, parse its parent directory.
				parseDir = filepath.Dir(filePath)
			}

			outputDir := c.String("output-directory")
			if outputDir != "" {
				if err := validateRelativePath(outputDir); err != nil {
					return fmt.Errorf("--output-directory: %w", err)
				}
			}

			// Progress, so it follows the per-action lines rather than
			// landing in the middle of the JSON document on stdout.
			_, _ = fmt.Fprintf(progress, "\nRegenerating YAML...\n")

			if err := p.Parse(provider.ProviderOps{
				Directory:       parseDir,
				OutputDirectory: outputDir,
				DryRun:          false,
			}); err != nil {
				return err
			}

			// The regenerated YAML is written either way: what upgraded is
			// correct, and holding it back over an action that did not would
			// leave the directory half converted. The failure is still the
			// command's outcome.
			return summaryErr
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "Upgrade actions in a single HCL `FILE`",
			},
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Value:   "",
				Usage:   "Upgrade actions in all HCL files in `DIRECTORY` (default: cinzel)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Value: false,
				Usage: "Show what would be upgraded without writing files",
			},
			&cli.BoolFlag{
				Name:  "check",
				Value: false,
				Usage: "Name the actions that are not on their latest version and exit non-zero, writing nothing",
			},
			&cli.BoolFlag{
				Name:  "json",
				Value: false,
				Usage: "Report the run as JSON on stdout, with progress on stderr",
			},
			&cli.StringFlag{
				Name:  "only-action",
				Value: "",
				Usage: "Upgrade only this `ACTION`, such as actions/checkout, leaving the rest alone",
			},
			&cli.IntFlag{
				Name:  "cooldown-days",
				Value: 0,
				Usage: "Leave a release alone until it has been published this many `DAYS`",
			},
			&cli.BoolFlag{
				Name:  "parse",
				Value: false,
				Usage: "Regenerate GitHub Actions YAML files after upgrading",
			},
			&cli.StringFlag{
				Name:  "output-directory",
				Value: "",
				Usage: "Output directory for parsed YAML (default: .github/workflows)",
			},
		},
	}
}

// printUpgradeSummary writes the counts and reports whether any action failed.
//
// Same as the pin side: the command returned nil over a printed "1 failed" and
// exited 0. Worse here, because --parse then regenerates YAML from a directory
// the run already knows it could not fully upgrade.
//
// Under check the upgraded count is the finding, for the reason given in
// printPinSummary, and a failure still outranks it.
func (cmd *Cli) printUpgradeSummary(results []pin.UpgradeResult, check bool) error {
	counts := countUpgrades(results)

	_, _ = fmt.Fprintf(cmd.Writer, "\nUpgrade summary: %d %s, %d already current, %d failed",
		counts.Upgraded, upgradedVerb(check), counts.AlreadyCurrent, counts.Failed)

	// Only when there are some. A run with no policy would otherwise carry a
	// "0 held in cooldown" it has no cooldown to hold anything in.
	if counts.Held > 0 {
		_, _ = fmt.Fprintf(cmd.Writer, ", %d held in cooldown", counts.Held)
	}

	// Same rule as the cooldown count: only when there are some, so a run
	// that targeted nothing does not carry a "0 skipped".
	if counts.Skipped > 0 {
		_, _ = fmt.Fprintf(cmd.Writer, ", %d not targeted", counts.Skipped)
	}

	_, _ = fmt.Fprintln(cmd.Writer)

	return upgradeOutcome(counts, len(results), check)
}

// upgradedVerb keeps the summary honest under check, where nothing was written.
func upgradedVerb(check bool) string {
	if check {
		return "out of date"
	}

	return "upgraded"
}

// upgradeCooldown reads the policy from the flag, or failing that from
// .cinzelrc.yaml.
//
// The flag wins when it is set, the way every other setting here resolves. The
// file is where it usually lives: a cooldown is a policy the repository keeps
// rather than something retyped on each run, and a scheduled upgrade job that
// has to remember the flag is one edit away from silently not having it.
func upgradeCooldown(c *cli.Command) (pin.Cooldown, error) {
	days := c.Int("cooldown-days")

	if !c.IsSet("cooldown-days") {
		configured, err := configuredCooldownDays()
		if err != nil {
			return pin.Cooldown{}, err
		}

		days = configured
	}

	if days < 0 {
		return pin.Cooldown{}, fmt.Errorf("%w: %d", errNegativeCooldown, days)
	}

	return pin.Cooldown{For: time.Duration(days) * 24 * time.Hour}, nil
}

// anyConsidered reports whether the run looked at any action at all.
//
// A targeted run skips everything it was not pointed at, and a skip is
// neither an upgrade nor a failure, so a name matching nothing produces a
// summary of all zeros that reads as success.
func anyConsidered(results []pin.UpgradeResult) bool {
	for _, r := range results {
		if !r.Skipped {
			return true
		}
	}

	return false
}
