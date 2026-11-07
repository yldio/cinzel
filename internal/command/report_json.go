// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yldio/cinzel/internal/pin"
)

// pinCounts is what a pin run found, and the summary of the JSON report.
type pinCounts struct {
	Pinned        int `json:"pinned"`
	AlreadyPinned int `json:"alreadyPinned"`
	Failed        int `json:"failed"`
}

// upgradeCounts is the same for upgrade. Held is counted apart from upgraded:
// a release inside its cooldown is neither taken nor a failure.
type upgradeCounts struct {
	Upgraded       int `json:"upgraded"`
	AlreadyCurrent int `json:"alreadyCurrent"`
	Failed         int `json:"failed"`
	Held           int `json:"held"`
}

// pinAction is one action in the JSON report.
//
// Status is the finding rather than the work, so it reads the same whether or
// not anything was written; dryRun on the envelope is what says which it was.
type pinAction struct {
	Action string `json:"action"`
	Status string `json:"status"`
	Tag    string `json:"tag,omitempty"`
	SHA    string `json:"sha,omitempty"`
	Error  string `json:"error,omitempty"`
}

// upgradeAction is one action in the upgrade JSON report.
type upgradeAction struct {
	Action     string `json:"action"`
	Status     string `json:"status"`
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	SHA        string `json:"sha,omitempty"`
	Published  string `json:"published,omitempty"`
	EligibleAt string `json:"eligibleAt,omitempty"`
	Error      string `json:"error,omitempty"`
}

type pinReport struct {
	Command string      `json:"command"`
	DryRun  bool        `json:"dryRun"`
	Summary pinCounts   `json:"summary"`
	Actions []pinAction `json:"actions"`
}

type upgradeReport struct {
	Command string          `json:"command"`
	DryRun  bool            `json:"dryRun"`
	Summary upgradeCounts   `json:"summary"`
	Actions []upgradeAction `json:"actions"`
}

// countPins classifies each result once, so the summary a reader sees and the
// exit code the run ends on cannot disagree.
func countPins(results []pin.PinResult) pinCounts {
	var c pinCounts

	for _, r := range results {
		switch {
		case r.WasAlready:
			c.AlreadyPinned++
		case r.Error != nil:
			c.Failed++
		default:
			c.Pinned++
		}
	}

	return c
}

func countUpgrades(results []pin.UpgradeResult) upgradeCounts {
	var c upgradeCounts

	for _, r := range results {
		switch {
		case r.WasCurrent:
			c.AlreadyCurrent++
		case r.Error != nil:
			c.Failed++
		case r.HeldFor > 0:
			c.Held++
		default:
			c.Upgraded++
		}
	}

	return c
}

// pinOutcome is the error a pin run ends on. Both the text and the JSON report
// call it, so the two cannot exit differently over the same results.
func pinOutcome(c pinCounts, total int, check bool) error {
	if c.Failed > 0 {
		return fmt.Errorf("%w: %d of %d", errPinFailed, c.Failed, total)
	}

	if check && c.Pinned > 0 {
		return fmt.Errorf("%w: %d of %d", errActionsNotPinned, c.Pinned, total)
	}

	return nil
}

func upgradeOutcome(c upgradeCounts, total int, check bool) error {
	if c.Failed > 0 {
		return fmt.Errorf("%w: %d of %d", errUpgradeFailed, c.Failed, total)
	}

	if check && c.Upgraded > 0 {
		return fmt.Errorf("%w: %d of %d", errActionsOutdated, c.Upgraded, total)
	}

	return nil
}

// printPinJSON writes the report and returns the same error the text summary
// would have.
//
// The failures are in the document rather than only on stderr, so a consumer
// that captured stdout has the whole run: reading why something could not be
// pinned should not mean parsing the human output as well.
func (cmd *Cli) printPinJSON(results []pin.PinResult, dryRun, check bool) error {
	report := pinReport{
		Command: "pin",
		DryRun:  dryRun,
		Summary: countPins(results),
		Actions: make([]pinAction, 0, len(results)),
	}

	for _, r := range results {
		action := pinAction{Action: r.Action, Tag: r.Tag, SHA: r.SHA}

		switch {
		case r.WasAlready:
			action.Status = "already-pinned"
		case r.Error != nil:
			action.Status = "failed"
			action.Error = r.Error.Error()
		default:
			action.Status = "pinned"
		}

		report.Actions = append(report.Actions, action)
	}

	if err := cmd.writeJSON(report); err != nil {
		return err
	}

	return pinOutcome(report.Summary, len(results), check)
}

func (cmd *Cli) printUpgradeJSON(results []pin.UpgradeResult, dryRun, check bool) error {
	report := upgradeReport{
		Command: "upgrade",
		DryRun:  dryRun,
		Summary: countUpgrades(results),
		Actions: make([]upgradeAction, 0, len(results)),
	}

	for _, r := range results {
		action := upgradeAction{
			Action: r.Action,
			From:   r.OldVersion,
			To:     r.NewTag,
			SHA:    r.NewSHA,
		}

		if !r.Published.IsZero() {
			action.Published = r.Published.UTC().Format(time.DateOnly)
		}

		switch {
		case r.WasCurrent:
			action.Status = "current"
		case r.Error != nil:
			action.Status = "failed"
			action.Error = r.Error.Error()
		case r.HeldFor > 0:
			action.Status = "held"

			// The date rather than the remaining duration, for the reason the
			// text report names it: "24d left" is a quantity with no noun on
			// it, and a consumer comparing against its own clock has to know
			// when the run happened to use it.
			if !r.EligibleAt.IsZero() {
				action.EligibleAt = r.EligibleAt.UTC().Format(time.DateOnly)
			}
		default:
			action.Status = "upgraded"
		}

		report.Actions = append(report.Actions, action)
	}

	if err := cmd.writeJSON(report); err != nil {
		return err
	}

	return upgradeOutcome(report.Summary, len(results), check)
}

// writeJSON writes one indented document and a closing newline.
//
// Indented because a person reads this too, often the first time they try the
// flag, and a terminal full of one line is no use to them. A consumer piping
// it into jq does not care either way.
func (cmd *Cli) writeJSON(report any) error {
	encoder := json.NewEncoder(cmd.Writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to write the JSON report: %w", err)
	}

	return nil
}
