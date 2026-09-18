// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"fmt"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

var (
	errInputPathRequired       = cinzelerror.UserInput(errors.New("`file` or `directory` must be set"))
	errInputPathConflict       = cinzelerror.UserInput(errors.New("`file` and `directory` cannot be set together"))
	errParseNotImplemented     = errors.New("gitlab parse is not implemented yet")
	errUnparseNotImplemented   = errors.New("gitlab unparse is not implemented yet")
	errBlockIDNotString        = cinzelerror.UserInput(errors.New("'id' must be a non-empty string"))
	errNoDefinitions           = cinzelerror.UserInput(errors.New("no pipeline definitions found in input"))
	errNoYAMLFiles             = cinzelerror.UserInput(errors.New("no YAML files found in input"))
	errYAMLExhausting          = cinzelerror.UserInput(errors.New("input would exhaust memory during YAML alias expansion"))
	errNonStringKey            = cinzelerror.UserInput(errors.New("a mapping key must be a string"))
	errInvalidUTF8             = cinzelerror.UserInput(errors.New("input is not valid UTF-8"))
	errNeedsJobEmpty           = cinzelerror.UserInput(errors.New("needs entries must name a job"))
	errJobNamedAfterKeyword    = cinzelerror.UserInput(errors.New("a job is named after a pipeline keyword"))
	errNeedJobNotSingle        = cinzelerror.UserInput(errors.New("a need block names one job"))
	errArtifactsNotAList       = cinzelerror.UserInput(errors.New("artifacts takes a single object"))
	errJobKeyReservedID        = cinzelerror.UserInput(errors.New("'id' is reserved: it records a job's name when the block label is sanitized"))
	errVariableKeyReservedName = cinzelerror.UserInput(errors.New("'name' is reserved: it records the variable's own name"))
	errJobIDHidden             = cinzelerror.UserInput(errors.New("a job 'id' cannot start with '.', which marks a hidden key GitLab never runs"))
	errNeedsHiddenJob          = cinzelerror.UserInput(errors.New("a hidden job never runs, so nothing can wait on one"))
	errReportsNotAList         = cinzelerror.UserInput(errors.New("reports takes a single object"))
	errExtendsNameEmpty        = cinzelerror.UserInput(errors.New("extends entries must name a job or template"))
	errTemplateIDDotted        = cinzelerror.UserInput(errors.New("a template 'id' is its key without the leading '.', which the reader puts back"))
)

// errUnknownKeyword reports a key the HCL schema does not declare, naming the
// block it was written in.
func errUnknownKeyword(owner, key string) error {
	return cinzelerror.UserInput(fmt.Errorf("unknown %s keyword '%s'", owner, key))
}

// errKeyNotAnIdentifier reports a passed-through top-level key that cannot be
// written as an HCL attribute name.
func errKeyNotAnIdentifier(key string) error {
	return cinzelerror.UserInput(fmt.Errorf("top-level key %q is not a valid HCL identifier, so it cannot be passed through", key))
}
