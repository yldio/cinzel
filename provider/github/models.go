// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

// WorkflowYAMLFile pairs a workflow filename with its YAML content and the
// order its jobs were declared in, which the emitter reproduces.
type WorkflowYAMLFile struct {
	Filename string
	Content  map[string]any
	JobOrder []string
	// FootComment closes the document, below its last key.
	FootComment string
	// Source is the HCL file this was declared in, spelled by fsutil.SourceKey.
	// It goes into the generated file's markers so a later run can tell which
	// outputs are its own to prune.
	Source string
}
