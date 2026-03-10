// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

// WorkflowYAMLFile pairs a workflow filename with its marshalled YAML content.
type WorkflowYAMLFile struct {
	Filename string
	Content  map[string]any
}
