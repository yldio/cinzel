// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

// PipelineYAMLFile pairs the output filename with pipeline YAML content.
type PipelineYAMLFile struct {
	Filename string
	Content  map[string]any
}
