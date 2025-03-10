// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package test

// HclBody is a stub type satisfying the Updater interface for HCL test fixtures.
type HclBody struct{}

func (h HclBody) Update(filename string) {}

// YamlBody is a stub type satisfying the Updater interface for YAML test fixtures.
type YamlBody struct {
	Name string `yaml:"name,omitempty"`
}

func (y YamlBody) Update(filename string) {}
