// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/hclparser"

	"github.com/zclconf/go-cty/cty"
)

// Parse resolves the HCL step configuration into a Step using the provided variables.
func (config *StepConfig) Parse(hv *hclparser.HCLVars) (Step, error) {

	if config == nil {

		return Step{}, nil
	}

	if config.Identifier == "" {

		return Step{}, fmt.Errorf("error in step: no identifier, %w", cinzelerror.ErrOpenIssue)
	}

	parsedStep := Step{
		Identifier: config.Identifier,
	}

	parsedIgnoreId, err := config.parseIgnoreId(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedIgnoreId != cty.NilVal {

		if err := parsedStep.parseIgnoreId(parsedIgnoreId); err != nil {

			return Step{}, err
		}
	}

	parsedId, err := config.parseId(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedIgnoreId != cty.True {

		if parsedId != cty.NilVal {

			if err := parsedStep.parseId(parsedId); err != nil {

				return Step{}, err
			}
		} else {
			// Keep backward compatibility with legacy behavior where step label
			// is used as the emitted step id unless explicitly ignored.

			if err := parsedStep.parseId(cty.StringVal(parsedStep.Identifier)); err != nil {
				parsedStep.Id = cty.NilVal
			}
		}
	}

	parsedIf, err := config.parseIf(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedIf != cty.NilVal {

		if err := parsedStep.parseIf(parsedIf); err != nil {

			return Step{}, err
		}
	}

	parsedName, err := config.parseName(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedName != cty.NilVal {

		if err := parsedStep.parseName(parsedName); err != nil {

			return Step{}, err
		}
	}

	parsedUses, err := config.parseUses(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedUses != cty.NilVal {

		if err := parsedStep.parseUses(parsedUses); err != nil {

			return Step{}, err
		}
	}

	parsedRun, err := config.parseRun(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedRun != cty.NilVal {

		if err := parsedStep.parseRun(parsedRun); err != nil {

			return Step{}, err
		}
	}

	parsedWorkingDirectory, err := config.parseWorkingDirectory(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedWorkingDirectory != cty.NilVal {

		if err := parsedStep.parseWorkingDirectory(parsedWorkingDirectory); err != nil {

			return Step{}, err
		}
	}

	parsedShell, err := config.parseShell(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedShell != cty.NilVal {

		if err := parsedStep.parseShell(parsedShell); err != nil {

			return Step{}, err
		}
	}

	parsedWith, err := config.parseWith(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedWith != cty.NilVal {

		if err := parsedStep.parseWith(parsedWith); err != nil {

			return Step{}, err
		}
	}

	parsedEnv, err := config.parseEnv(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedEnv != cty.NilVal {

		if err := parsedStep.parseEnv(parsedEnv); err != nil {

			return Step{}, err
		}
	}

	parsedContinueOnError, err := config.parseContinueOnError(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedContinueOnError != cty.NilVal {

		if err := parsedStep.parseContinueOnError(parsedContinueOnError); err != nil {

			return Step{}, err
		}
	}

	parsedTimeoutMinutes, err := config.parseTimeoutMinutes(hv)
	if err != nil {

		return Step{}, fmt.Errorf("error in step '%s': %w, %w", parsedStep.Identifier, err, cinzelerror.ErrOpenIssue)
	}

	if parsedTimeoutMinutes != cty.NilVal {

		if err := parsedStep.parseTimeoutMinutes(parsedTimeoutMinutes); err != nil {

			return Step{}, err
		}
	}

	return parsedStep, nil
}

// Parse resolves all step configurations in the list into a Steps map.
func (config *StepListConfig) Parse(hv *hclparser.HCLVars) (Steps, error) {
	steps := make(Steps)

	for _, step := range *config {
		parsedStep, err := step.Parse(hv)
		if err != nil {

			return Steps{}, err
		}

		_, exists := steps[parsedStep.Identifier]

		if exists {

			return Steps{}, fmt.Errorf("error in step '%s': already defined, %w", parsedStep.Identifier, cinzelerror.ErrOpenIssue)
		}

		steps[parsedStep.Identifier] = parsedStep
	}

	return steps, nil
}
