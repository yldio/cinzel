// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func errOnlyHclFiles() error { return errors.New("only HCL files are allowed") }

func errOnRestriction() error { return errors.New("`on` can only have Events or Event") }

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
func ProcessHCLDiags(diags hcl.Diagnostics) error {
	errs := make([]error, 0, len(diags))

	for _, diag := range diags {

		if diag.Detail != "" {
			errs = append(errs, errors.New(diag.Detail))
		}
	}

	return fmt.Errorf("%w, %w", errors.Join(errs...), ErrOpenIssue)
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
