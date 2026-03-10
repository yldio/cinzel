// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package test

// HclBody is a stub type satisfying the Updater interface for HCL test fixtures.
type HclBody struct{}

// Update is a no-op implementation used by HCL fixture tests.
func (h HclBody) Update(filename string) {}

// YamlBody is a stub type satisfying the Updater interface for YAML test fixtures.
type YamlBody struct {
	Name string `yaml:"name,omitempty"`
}

// Update is a no-op implementation used by YAML fixture tests.
func (y YamlBody) Update(filename string) {}
