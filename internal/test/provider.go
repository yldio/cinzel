// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package test

import (
	"errors"
	"io"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// MockProvider returns a mock provider for use in tests.
func MockProvider(t *testing.T, writer io.Writer) *mockProvider {
	t.Helper()

	return &mockProvider{
		sink: writer,
	}
}

type mockProvider struct {
	sink     io.Writer
	HasError bool
}

// Parse writes a parse marker unless error mode is enabled.
func (p *mockProvider) Parse(opts provider.ProviderOps) error {
	if p.HasError {
		return errors.New("parse error")
	}

	_, err := p.sink.Write([]byte("parse"))

	return err
}

// Unparse writes an unparse marker unless error mode is enabled.
func (p *mockProvider) Unparse(opts provider.ProviderOps) error {
	if p.HasError {
		return errors.New("unparse error")
	}

	_, err := p.sink.Write([]byte("unparse"))

	return err
}

// GetProviderName returns the mock provider name.
func (p *mockProvider) GetProviderName() string {
	return "mock-provider"
}

// GetDescription returns the mock provider description.
func (p *mockProvider) GetDescription() string {
	return "mock-provider description"
}

// GetParseDescription returns the mock parse description.
func (p *mockProvider) GetParseDescription() string {
	return "mock-provider parse"
}

// GetUnparseDescription returns the mock unparse description.
func (p *mockProvider) GetUnparseDescription() string {
	return "mock-provider unparse"
}
