// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

// WorkflowYAMLFile pairs a workflow filename with its YAML content and the
// order its jobs were declared in, which the emitter reproduces.
type WorkflowYAMLFile struct {
	Filename string
	Content  map[string]any
	JobOrder []string
}
