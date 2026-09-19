// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

var (
	errInputPathRequired      = cinzelerror.UserInput(errors.New("`file` or `directory` must be set"))
	errInputPathConflict      = cinzelerror.UserInput(errors.New("`file` and `directory` cannot be set together"))
	errNoYAMLFiles            = cinzelerror.UserInput(errors.New("no YAML files found in input"))
	errNoDefinitions          = cinzelerror.UserInput(errors.New("no workflow, action or step definitions found in input"))
	errUnsupportedBodyType    = errors.New("unsupported body type")
	errUnsupportedBlockBody   = errors.New("unsupported block body type")
	errUnsupportedUsesBody    = errors.New("unsupported uses block body type")
	errNamedBlockMissingName  = cinzelerror.UserInput(errors.New("block must include a 'name' attribute"))
	errNamedBlockMissingValue = cinzelerror.UserInput(errors.New("block must include a 'value' attribute"))
	errWorkflowYAMLOnJobs     = cinzelerror.UserInput(errors.New("workflow YAML must define both 'on' and 'jobs'"))
	errJobIDNotString         = cinzelerror.UserInput(errors.New("job 'id' must be a non-empty string"))
	errMultipleDocuments      = cinzelerror.UserInput(errors.New("a workflow file must hold a single YAML document"))
	errNonStringKey           = cinzelerror.UserInput(errors.New("a mapping key must be a string"))
	errFilenameEscapes        = cinzelerror.UserInput(errors.New("'filename' must stay inside the output directory"))
	errDuplicateFilename      = cinzelerror.UserInput(errors.New("two definitions write to the same file"))
	errDuplicateJobLabel      = cinzelerror.UserInput(errors.New("two job blocks share a label"))
	errDuplicateJobKey        = cinzelerror.UserInput(errors.New("two jobs write to the same YAML key"))
	errDuplicateStepID        = cinzelerror.UserInput(errors.New("two steps in one job write the same step id"))
	errDuplicateActionStepID  = cinzelerror.UserInput(errors.New("two steps in one action write the same step id"))
	errDuplicateStepLabel     = cinzelerror.UserInput(errors.New("two step blocks share a label"))
	errDuplicateBlockKey      = cinzelerror.UserInput(errors.New("two blocks write the same key"))
	errDuplicateBlock         = cinzelerror.UserInput(errors.New("a block that may only be written once is written twice"))
	errEmptyStep              = cinzelerror.UserInput(errors.New("a step a job or an action runs must set something"))
	errEmptyBlock             = cinzelerror.UserInput(errors.New("a block that becomes a YAML key must set something"))
	errEmptyKey               = cinzelerror.UserInput(errors.New("a key a workflow reads by name must not be empty"))
	errInfiniteNumber         = cinzelerror.UserInput(errors.New("an infinite number cannot be written to HCL"))
	errNeedsOutsideWorkflow   = cinzelerror.UserInput(errors.New("a job can only wait on a job the same workflow writes"))
)
