// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestName(t *testing.T) {
	type Test struct {
		name   string
		have   *NameConfig
		expect string
	}

	var name = "job_name"

	var have1 = NameConfig(name)
	var expect1 = name

	var have2 = NameConfig("")
	var expect2 = ""

	var expect3 = ""

	var tests = []Test{
		{"with defined name", &have1, expect1},
		{"without empty name", &have2, expect2},
		{"without undefined name", nil, expect3},
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
