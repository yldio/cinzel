// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestIf(t *testing.T) {
	type Test struct {
		name   string
		have   *IfConfig
		expect string
	}

	var iF = "${{ ! startsWith(github.ref, 'refs/tags/') }}"

	var have_1 = IfConfig(iF)
	var expect_1 = iF

	var have_2 = IfConfig("")
	var expect_2 = ""

	var expect_3 = ""

	var tests = []Test{
		{"with defined if", &have_1, expect_1},
		{"without empty if", &have_2, expect_2},
		{"without undefined if", nil, expect_3},
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
