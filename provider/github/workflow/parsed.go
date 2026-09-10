// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package workflow

// Parsed holds the intermediate representation of a workflow after HCL parsing.
type Parsed struct {
	ID       string
	Filename string
	Body     map[string]any
	JobRefs  []string
}

// ValidationModel contains the fields needed to validate a workflow definition.
type ValidationModel struct {
	ID      string
	HasOn   bool
	OnCount int
	JobRefs []string
}
