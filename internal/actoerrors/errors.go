// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package actoerrors

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

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

func ErrWorkflowEmptyJobs(workflowId string) error {
	return fmt.Errorf("workflow `%s` requires at least one job", workflowId)
}

func errWorkflowEmptyOn() error {
	return errors.New("has to have at leat one `on` event")
}

func ErrJobEmptySteps(jobId string) error {
	return fmt.Errorf("job `%s` requires at least one `step`", jobId)
}

func ProcessHCLDiags(diags hcl.Diagnostics) error {
	var err error

	for _, diag := range diags {
		err = fmt.Errorf("%s", diag.Detail)
	}

	return fmt.Errorf("%w, %w", err, ErrOpenIssue)
}

func errOpenIssue() error {
	return errors.New("if you think this is incorrect, consider opening an issue in https://www.github.com/yldio/acto/issues")
}
