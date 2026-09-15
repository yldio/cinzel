// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import "errors"

var (
	errInputPathRequired     = errors.New("`file` or `directory` must be set")
	errInputPathConflict     = errors.New("`file` and `directory` cannot be set together")
	errParseNotImplemented   = errors.New("gitlab parse is not implemented yet")
	errUnparseNotImplemented = errors.New("gitlab unparse is not implemented yet")
	errBlockIDNotString      = errors.New("'id' must be a non-empty string")
	errNoDefinitions         = errors.New("no pipeline definitions found in input")
	errNoYAMLFiles           = errors.New("no YAML files found in input")
	errYAMLExhausting        = errors.New("input would exhaust memory during YAML alias expansion")
	errNonStringKey          = errors.New("a mapping key must be a string")
	errInvalidUTF8           = errors.New("input is not valid UTF-8")
)
