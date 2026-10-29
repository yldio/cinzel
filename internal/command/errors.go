// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"errors"
	"fmt"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

var (
	errCancelled        = cinzelerror.UserInput(errors.New("cancelled"))
	errPromptRequired   = cinzelerror.UserInput(errors.New("--prompt is required (or use --refine to iterate on previous output)"))
	errAbsolutePath     = cinzelerror.UserInput(errors.New("path must be relative to the project directory"))
	errPinFailed        = cinzelerror.UserInput(errors.New("some actions could not be pinned"))
	errUpgradeFailed    = cinzelerror.UserInput(errors.New("some actions could not be upgraded"))
	errPathTraversal    = cinzelerror.UserInput(errors.New("path must not escape the project directory (no .. traversal)"))
	errFileIsADirectory = cinzelerror.UserInput(errors.New("`file` names a directory, use `directory` for that"))

	// ErrDiffFound is returned when the generated files are not what the HCL
	// would produce. It is the whole point of the diff command rather than a
	// fault, so it exits 1 and says only that, with the files already named.
	ErrDiffFound = cinzelerror.UserInput(errors.New("the generated files are not what the HCL produces"))

	// ErrCannotRun marks a diff that could not reach an answer: unreadable
	// HCL, a missing directory, a provider that cannot diff.
	//
	// It exists so a workflow can tell "the YAML drifted" from "cinzel did not
	// run", which are the same red step otherwise. Only diff carries it, so
	// every other command keeps exiting 1 exactly as before.
	ErrCannotRun = errors.New("cinzel could not work out what parse would change")

	// errUnformatted ends a "fmt --check" run that found a file the
	// formatter would rewrite. The files are already named on the way
	// through, so the error says only how many there were.
	errUnformatted = cinzelerror.UserInput(errors.New("some HCL files are not formatted"))

	errDiffUnsupported        = cinzelerror.UserInput(errors.New("this provider cannot report what parse would change"))
	errRootAndOutputDirectory = cinzelerror.UserInput(errors.New("--root and --output-directory cannot be set together: --root moves the whole run, --output-directory only the files it names"))
)

// cannotRun marks err as a diff that never got to an answer, which main reads
// as exit 2.
//
// A wrapped error that was already the author's to fix stays that way: wrapping
// alone would leave IsUserInput looking at an unmarked ErrCannotRun and put the
// open-an-issue line under a misspelled directory.
func cannotRun(err error) error {
	wrapped := fmt.Errorf("%w: %w", ErrCannotRun, err)

	if cinzelerror.IsUserInput(err) {
		return cinzelerror.UserInput(wrapped)
	}

	return wrapped
}
