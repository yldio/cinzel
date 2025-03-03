// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package provider

type ProviderOps struct {
	File            string
	Directory       string
	OutputDirectory string
	Recursive       bool
	DryRun          bool
	Override        bool
	Watch           bool
}

type Provider interface {
	Parse(opts ProviderOps) error
	Unparse(opts ProviderOps) error
	GetProviderName() string
	GetDescription() string
	GetParseDescription() string
	GetUnparseDescription() string
	DefaultOutputDirectory() string
}
