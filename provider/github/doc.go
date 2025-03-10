// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package github implements the Provider interface for GitHub Actions.
// It converts between HCL definitions and GitHub Actions YAML (workflows
// and composite actions), supporting bidirectional parse (HCL→YAML) and
// unparse (YAML→HCL) flows.
package github
