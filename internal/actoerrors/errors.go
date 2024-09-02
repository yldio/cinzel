// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

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

func ErrWorkflowEmptyOn(workflowId string) error {
	return fmt.Errorf("workflow `%s` requires at least one `on` event", workflowId)
}

func ErrJobEmptySteps(jobId string) error {
	return fmt.Errorf("job `%s` requires at least one `step`", jobId)
}

func ProcessHCLDiags(diags hcl.Diagnostics) error {
	var errorsList []error

	for _, diag := range diags {
		err := fmt.Errorf("%s on %s", diag.Summary, diag.Subject)
		errorsList = append(errorsList, err)
	}

	return errors.Join(errorsList...)
}
