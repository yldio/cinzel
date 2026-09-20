// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/hclparser"

	"github.com/zclconf/go-cty/cty"
)

// stepAttrErr names the step and the attribute a value was refused for.
//
// A type check reports the type it wanted and the one it found and nothing
// else, so "unsupported type, expected string, found number" left an author
// with four string attributes on a dozen steps nowhere to look. The
// expression errors beside them already name the step.
func stepAttrErr(id, attr string, err error) error {
	return fmt.Errorf("error in step '%s': %s: %w", id, attr, err)
}

// stepExprErr names the step an expression was refused in.
//
// The open-an-issue line used to be written in at each of these by hand, which
// put it on an error the author caused: "variable.list_os[\"prod\"]" is their
// own index to fix, and cinzelerror.New leaves the line off a marked error for
// exactly that reason. Letting the error through unwrapped lets the mark it
// already carries decide, the same way the job path does.
func stepExprErr(id string, err error) error {
	return fmt.Errorf("error in step '%s': %w", id, err)
}

// Parse resolves the HCL step configuration into a Step using the provided variables.
func (config *StepConfig) Parse(hv *hclparser.HCLVars) (Step, error) {
	if config == nil {
		return Step{}, nil
	}

	if config.Identifier == "" {
		return Step{}, fmt.Errorf("error in step: no identifier, %w", cinzelerror.ErrOpenIssue)
	}

	if err := hclparser.RejectUnknown(config.Body); err != nil {
		return Step{}, fmt.Errorf("error in step '%s': %w", config.Identifier, err)
	}

	head, foot := blockComments(config.Body, hv)

	parsedStep := Step{
		Identifier: config.Identifier,
		Comments:   Comments{Head: head, Foot: foot},
	}

	parsedIgnoreId, err := config.parseIgnoreId(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedIgnoreId != cty.NilVal {
		if err := parsedStep.parseIgnoreId(parsedIgnoreId); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "ignore_id", err)
		}
	}

	parsedId, err := config.parseId(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedIgnoreId != cty.True {
		if parsedId != cty.NilVal {
			if err := parsedStep.parseId(parsedId); err != nil {
				return Step{}, stepAttrErr(parsedStep.Identifier, "id", err)
			}

			parsedStep.Comments.setExpr(hv, "id", config.Id, parsedId)
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
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedIf != cty.NilVal {
		if err := parsedStep.parseIf(parsedIf); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "if", err)
		}

		parsedStep.Comments.setExpr(hv, "if", config.If, parsedIf)
	}

	parsedName, err := config.parseName(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedName != cty.NilVal {
		if err := parsedStep.parseName(parsedName); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "name", err)
		}

		parsedStep.Comments.setExpr(hv, "name", config.Name, parsedName)
	}

	parsedUses, usesComment, err := config.parseUses(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedUses != cty.NilVal {
		if err := parsedStep.parseUses(parsedUses); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "uses", err)
		}

		// The uses block's comment is written on its "version" attribute,
		// which is where the pin tag lands, and the whole block becomes the
		// single "uses" key.
		if usesComment != "" {
			if parsedStep.Comments.Attrs == nil {
				parsedStep.Comments.Attrs = map[string]Comment{}
			}

			parsedStep.Comments.Attrs["uses"] = Comment{Line: usesComment}
		}
	}

	parsedRun, err := config.parseRun(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedRun != cty.NilVal {
		if err := parsedStep.parseRun(parsedRun); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "run", err)
		}

		parsedStep.Comments.setExpr(hv, "run", config.Run, parsedRun)
	}

	parsedWorkingDirectory, err := config.parseWorkingDirectory(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedWorkingDirectory != cty.NilVal {
		if err := parsedStep.parseWorkingDirectory(parsedWorkingDirectory); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "working_directory", err)
		}

		parsedStep.Comments.setExpr(hv, "working-directory", config.WorkingDirectory, parsedWorkingDirectory)
	}

	parsedShell, err := config.parseShell(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedShell != cty.NilVal {
		if err := parsedStep.parseShell(parsedShell); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "shell", err)
		}

		parsedStep.Comments.setExpr(hv, "shell", config.Shell, parsedShell)
	}

	parsedWith, err := config.parseWith(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedWith != cty.NilVal {
		if err := parsedStep.parseWith(parsedWith); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "with", err)
		}

		parsedStep.Comments.setNested(hv, "with", config.With.ValueRanges(hv))
	}

	parsedEnv, err := config.parseEnv(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedEnv != cty.NilVal {
		if err := parsedStep.parseEnv(parsedEnv); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "env", err)
		}

		parsedStep.Comments.setNested(hv, "env", config.Env.ValueRanges(hv))
	}

	parsedContinueOnError, err := config.parseContinueOnError(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedContinueOnError != cty.NilVal {
		if err := parsedStep.parseContinueOnError(parsedContinueOnError); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "continue_on_error", err)
		}

		parsedStep.Comments.setExpr(hv, "continue-on-error", config.ContinueOnError, parsedContinueOnError)
	}

	parsedTimeoutMinutes, err := config.parseTimeoutMinutes(hv)
	if err != nil {
		return Step{}, stepExprErr(parsedStep.Identifier, err)
	}

	if parsedTimeoutMinutes != cty.NilVal {
		if err := parsedStep.parseTimeoutMinutes(parsedTimeoutMinutes); err != nil {
			return Step{}, stepAttrErr(parsedStep.Identifier, "timeout_minutes", err)
		}

		parsedStep.Comments.setExpr(hv, "timeout-minutes", config.TimeoutMinutes, parsedTimeoutMinutes)
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
