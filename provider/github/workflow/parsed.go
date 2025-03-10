// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

// Parsed holds the intermediate representation of a workflow after HCL parsing.
type Parsed struct {
	ID       string
	Filename string
	Body     map[string]any
	JobRefs  []string
}

// NewParsed creates a Parsed workflow, extracting filename and job references from the body.
func NewParsed(id string, body map[string]any) Parsed {
	w := Parsed{ID: id, Body: body}
	if filename, ok := body["filename"].(string); ok {
		w.Filename = filename
	}
	delete(w.Body, "filename")

	if refs, ok := body["jobsRefs"].([]string); ok {
		w.JobRefs = refs
		delete(w.Body, "jobsRefs")
	}

	return w
}

// ValidationModel contains the fields needed to validate a workflow definition.
type ValidationModel struct {
	ID      string
	HasOn   bool
	OnCount int
	JobRefs []string
}
