// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package provider

// ProviderOps holds the options passed to a provider's Parse or Unparse operation.
type ProviderOps struct {
	File            string
	Directory       string
	OutputDirectory string
	Recursive       bool
	DryRun          bool
	YML             bool // use .yml extension instead of .yaml
}

// Provider defines the interface that each CI/CD provider must implement.
type Provider interface {
	Parse(opts ProviderOps) error
	Unparse(opts ProviderOps) error
	GetProviderName() string
	GetDescription() string
	GetParseDescription() string
	GetUnparseDescription() string
}
