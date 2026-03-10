// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

// PipelineYAMLFile pairs the output filename with pipeline YAML content.
type PipelineYAMLFile struct {
	Filename string
	Content  map[string]any
}
