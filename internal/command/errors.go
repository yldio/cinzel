// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"errors"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

var (
	errCancelled      = cinzelerror.UserInput(errors.New("cancelled"))
	errPromptRequired = cinzelerror.UserInput(errors.New("--prompt is required (or use --refine to iterate on previous output)"))
	errAbsolutePath   = cinzelerror.UserInput(errors.New("path must be relative to the project directory"))
	errPinFailed      = cinzelerror.UserInput(errors.New("some actions could not be pinned"))
	errUpgradeFailed  = cinzelerror.UserInput(errors.New("some actions could not be upgraded"))
	errPathTraversal  = cinzelerror.UserInput(errors.New("path must not escape the project directory (no .. traversal)"))
)
