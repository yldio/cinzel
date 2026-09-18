// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"fmt"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

var (
	errInputPathRequired     = cinzelerror.UserInput(errors.New("`file` or `directory` must be set"))
	errInputPathConflict     = cinzelerror.UserInput(errors.New("`file` and `directory` cannot be set together"))
	errParseNotImplemented   = errors.New("gitlab parse is not implemented yet")
	errUnparseNotImplemented = errors.New("gitlab unparse is not implemented yet")
	errBlockIDNotString      = cinzelerror.UserInput(errors.New("'id' must be a non-empty string"))
	errNoDefinitions         = cinzelerror.UserInput(errors.New("no pipeline definitions found in input"))
	errNoYAMLFiles           = cinzelerror.UserInput(errors.New("no YAML files found in input"))
	errYAMLExhausting        = cinzelerror.UserInput(errors.New("input would exhaust memory during YAML alias expansion"))
	errNonStringKey          = cinzelerror.UserInput(errors.New("a mapping key must be a string"))
	errInvalidUTF8           = cinzelerror.UserInput(errors.New("input is not valid UTF-8"))
	errNeedsJobEmpty         = cinzelerror.UserInput(errors.New("needs entries must name a job"))
	errJobNamedAfterKeyword  = cinzelerror.UserInput(errors.New("a job is named after a pipeline keyword"))
	errNeedJobNotSingle      = cinzelerror.UserInput(errors.New("a need block names one job"))
	errArtifactsNotAList     = cinzelerror.UserInput(errors.New("artifacts takes a single object"))
	errReportsNotAList       = cinzelerror.UserInput(errors.New("reports takes a single object"))
)

// errKeyNotAnIdentifier reports a passed-through top-level key that cannot be
// written as an HCL attribute name.
func errKeyNotAnIdentifier(key string) error {
	return cinzelerror.UserInput(fmt.Errorf("top-level key %q is not a valid HCL identifier, so it cannot be passed through", key))
}
