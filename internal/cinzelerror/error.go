// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package cinzelerror

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// Sentinel errors for workflow and HCL validation.
var (
	ErrWorkflowFilenameRequired = errWorkflowFilenameRequired()
	ErrOnlyHclFiles             = errOnlyHclFiles()
	ErrOnRestriction            = errOnRestriction()
	ErrSecretsRestriction       = errSecretsRestriction()
	ErrWorkflowEmptyOn          = errWorkflowEmptyOn()
	ErrOpenIssue                = errOpenIssue()
)

func errWorkflowFilenameRequired() error { return errors.New("`workflow` requires a filename") }
func errOnlyHclFiles() error             { return errors.New("only HCL files are allowed") }
func errOnRestriction() error            { return errors.New("`on` can only have Events or Event") }

func errSecretsRestriction() error {
	return errors.New("only `secrets` blocks or one single `secret` attribute is allowed")
}

// ErrWorkflowEmptyJobs returns an error indicating the workflow has no jobs.
func ErrWorkflowEmptyJobs(workflowId string) error {
	return fmt.Errorf("workflow `%s` requires at least one job", workflowId)
}

func errWorkflowEmptyOn() error {
	return errors.New("has to have at least one `on` event")
}

// ErrJobEmptySteps returns an error indicating the job has no steps.
func ErrJobEmptySteps(jobId string) error {
	return fmt.Errorf("job `%s` requires at least one `step`", jobId)
}

// ProcessHCLDiags converts HCL diagnostics into a single joined error.
//
// A diagnostic's summary stands in when it carries no detail. Only the detail
// used to be collected, so a summary-only diagnostic left nothing to join and
// errors.Join returned nil, which the caller printed as "%!w(<nil>)" with the
// report-an-issue line after it and no sign of what was actually wrong.
func ProcessHCLDiags(diags hcl.Diagnostics) error {
	errs := make([]error, 0, len(diags))

	for _, diag := range diags {
		message := diag.Detail

		if message == "" {
			message = diag.Summary
		}

		if message != "" {
			errs = append(errs, errors.New(message))
		}
	}

	// Diagnostics that carry no message at all leave nothing to report, and
	// that is cinzel's problem rather than the author's.
	if len(errs) == 0 {
		return ErrOpenIssue
	}

	// A diagnostic describes what is wrong with what was written, so it is
	// reported without the invitation to file a bug.
	return UserInput(errors.Join(errs...))
}

func errOpenIssue() error {
	return errors.New("if you think this is incorrect, consider opening an issue in https://www.github.com/yldio/cinzel/issues")
}

// OpenIssue is the message appended to errors suggesting users file a bug report.
const (
	OpenIssue string = "if you think this is incorrect, consider opening an issue in https://www.github.com/yldio/cinzel/issues"
)

// Error wraps an underlying error with additional context and issue-reporting guidance.
type Error struct {
	Err error
}

// New creates an Error from err and optional context messages, appending the OpenIssue text.
func New(err error, messages ...string) Error {
	parts := make([]string, 0, len(messages))

	for _, m := range messages {
		if m != "" {
			parts = append(parts, m)
		}
	}
	prefix := strings.Join(parts, ", ")

	if err != nil {
		// What the input caused is the author's to fix, so it is reported
		// without the invitation to file a bug.
		if IsUserInput(err) {
			if prefix != "" {
				return Error{Err: fmt.Errorf("%s: %w", prefix, err)}
			}

			return Error{Err: err}
		}

		// Several errors are built with the suffix already on them, either
		// through ErrOpenIssue or through ProcessHCLDiags. Adding a second
		// copy here printed the same sentence twice.
		if strings.Contains(err.Error(), OpenIssue) {
			if prefix != "" {
				return Error{Err: fmt.Errorf("%s: %w", prefix, err)}
			}

			return Error{Err: err}
		}

		if prefix != "" {
			return Error{Err: fmt.Errorf("%s: %w, %s", prefix, err, OpenIssue)}
		}

		return Error{Err: fmt.Errorf("%w, %s", err, OpenIssue)}
	}

	if prefix != "" {
		return Error{Err: fmt.Errorf("%s: %s", prefix, OpenIssue)}
	}

	return Error{Err: fmt.Errorf("%s", OpenIssue)}
}

// NewFromResource creates an Error annotated with the resource type and identifier.
func NewFromResource(err error, resourceType string, resourceId string) Error {
	var message string

	if resourceType != "" {
		message = fmt.Sprintf("error in %s", resourceType)

		if resourceId != "" {
			message = fmt.Sprintf("%s '%s'", message, resourceId)
		}
	}

	return New(err, message)
}

// Error returns the string representation of the wrapped error.
func (e *Error) Error() string { return e.Err.Error() }

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error { return e.Err }

// userInputError marks an error the input caused, as against one cinzel is at
// fault for.
type userInputError struct{ err error }

func (e userInputError) Error() string { return e.err.Error() }
func (e userInputError) Unwrap() error { return e.err }

// UserInput marks err as caused by what was written, not by a defect in
// cinzel. New leaves the open-an-issue line off these: a duplicate label or a
// misspelled attribute is the author's to fix, and pointing them at the issue
// tracker over it wasted their time and ours.
//
// Wrapping a sentinel at its declaration carries the mark through every
// fmt.Errorf("%w") built on it, and leaves errors.Is comparisons against that
// sentinel working unchanged.
func UserInput(err error) error { return userInputError{err: err} }

// IsUserInput reports whether err, or anything it wraps, is marked.
func IsUserInput(err error) bool {
	var target userInputError

	return errors.As(err, &target)
}
