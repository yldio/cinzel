// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package actoerrors

import "errors"

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
