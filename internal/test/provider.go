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

func (p *mockProvider) Parse(opts provider.ProviderOps) error {

	if p.HasError {

		return errors.New("parse error")
	}

	_, err := p.sink.Write([]byte("parse"))

	return err
}

func (p *mockProvider) Unparse(opts provider.ProviderOps) error {

	if p.HasError {

		return errors.New("unparse error")
	}

	_, err := p.sink.Write([]byte("unparse"))

	return err
}

func (p *mockProvider) GetProviderName() string {

	return "mock-provider"
}

func (p *mockProvider) GetDescription() string {

	return "mock-provider description"
}

func (p *mockProvider) GetParseDescription() string {

	return "mock-provider parse"
}

func (p *mockProvider) GetUnparseDescription() string {

	return "mock-provider unparse"
}
