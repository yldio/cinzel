// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package step

import (
	"reflect"
	"testing"
)

func TestSteps(t *testing.T) {
	type Test struct {
		name   string
		have   *StepsConfig
		expect Steps
	}

	var have1 = StepsConfig{
		{
			Id: "step1",
		},
	}
	var expect1 = Steps{
		{
			Id: "step1",
		},
	}

	var tests = []Test{
		{"with defined Id", &have1, expect1},
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
