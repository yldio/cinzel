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

	var have1 = IfConfig(iF)
	var expect1 = iF

	var have2 = IfConfig("")
	var expect2 = ""

	var expect3 = ""

	var tests = []Test{
		{"with defined if", &have1, expect1},
		{"without empty if", &have2, expect2},
		{"without undefined if", nil, expect3},
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
