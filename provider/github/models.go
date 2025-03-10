// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

// WorkflowYAMLFile pairs a workflow filename with its marshalled YAML content.
type WorkflowYAMLFile struct {
	Filename string
	Content  map[string]any
}
