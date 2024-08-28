// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestConcurrency(t *testing.T) {
	type Test struct {
		name   string
		have   *ConcurrencyConfig
		expect Concurrency
	}

	var group1 = "${{ github.workflow }}-${{ github.ref }}"
	var cancelInProgress1 = true
	var group2 = "${{ github.workflow }}-${{ github.ref }}"

	var have1 = ConcurrencyConfig{
		Group:            &group1,
		CancelInProgress: &cancelInProgress1,
	}
	var expect1 = Concurrency{
		Group:            &group1,
		CancelInProgress: &cancelInProgress1,
	}

	var have2 = ConcurrencyConfig{
		Group: &group2,
	}
	var expect2 = Concurrency{
		Group: &group2,
	}

	var tests = []Test{
		{"with defined concurrency, group and Cancel-in-progress", &have1, expect1},
		{"with defined concurrency and group", &have2, expect2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Error(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}
}
