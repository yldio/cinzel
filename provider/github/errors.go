// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import "errors"

var (
	errInputPathRequired      = errors.New("`file` or `directory` must be set")
	errInputPathConflict      = errors.New("`file` and `directory` cannot be set together")
	errNoYAMLFiles            = errors.New("no YAML files found in input")
	errUnsupportedBodyType    = errors.New("unsupported body type")
	errUnsupportedBlockBody   = errors.New("unsupported block body type")
	errUnsupportedUsesBody    = errors.New("unsupported uses block body type")
	errNamedBlockMissingName  = errors.New("block must include a 'name' attribute")
	errNamedBlockMissingValue = errors.New("block must include a 'value' attribute")
	errWorkflowYAMLOnJobs     = errors.New("workflow YAML must define both 'on' and 'jobs'")
)
