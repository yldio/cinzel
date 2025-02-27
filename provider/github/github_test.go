// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package github

import (
	"testing"
)

func TestGitHub(t *testing.T) {
	type test struct {
		name   string
		have   []byte
		expect []byte
	}

	var tests = []test{
		{"test", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
