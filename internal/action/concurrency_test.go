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

	var group_1 = "${{ github.workflow }}-${{ github.ref }}"
	var cancelInProgress_1 = true
	var group_2 = "${{ github.workflow }}-${{ github.ref }}"

	var have_1 = ConcurrencyConfig{
		Group:            &group_1,
		CancelInProgress: &cancelInProgress_1,
	}
	var expect_1 = Concurrency{
		Group:            &group_1,
		CancelInProgress: &cancelInProgress_1,
	}

	var have_2 = ConcurrencyConfig{
		Group: &group_2,
	}
	var expect_2 = Concurrency{
		Group: &group_2,
	}

	var tests = []Test{
		{"with defined concurrency, group and Cancel-in-progress", &have_1, expect_1},
		{"with defined concurrency and group", &have_2, expect_2},
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
