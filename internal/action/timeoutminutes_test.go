// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestTimeoutMinutes(t *testing.T) {
	type Test struct {
		name   string
		have   *TimeoutMinutesConfig
		expect *uint16
	}

	var have1 = TimeoutMinutesConfig(5)
	var number1 = uint16(5)
	var expect1 = &number1

	var tests = []Test{
		{"with defined timeout-minutes", &have1, expect1},
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
