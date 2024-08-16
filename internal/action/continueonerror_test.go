// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestContinueOnError(t *testing.T) {
	type Test struct {
		name   string
		have   *ContinueOnErrorConfig
		expect *bool
	}

	var have_1 = ContinueOnErrorConfig(true)
	var boolean = true
	var expect_1 = &boolean

	var tests = []Test{
		{"with defined continue-on-error", &have_1, expect_1},
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
