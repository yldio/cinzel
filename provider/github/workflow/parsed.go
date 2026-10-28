// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package workflow

// Parsed holds the intermediate representation of a workflow after HCL parsing.
type Parsed struct {
	ID       string
	Filename string
	Body     map[string]any
	JobRefs  []string
	// JobsComment holds what was written above the "jobs =" list and beside
	// it. The list is a reference to blocks written elsewhere, so it has no
	// value of its own to carry a comment: the "jobs" key it becomes does.
	JobsComment Comment
}

// Comment holds the comments written above a reference list and beside it.
type Comment struct {
	Head string
	Line string
}

// ValidationModel contains the fields needed to validate a workflow definition.
type ValidationModel struct {
	ID      string
	HasOn   bool
	OnCount int
	JobRefs []string
}
