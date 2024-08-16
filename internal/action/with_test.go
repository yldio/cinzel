// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestWith(t *testing.T) {
	type Test struct {
		name   string
		have   *WithsConfig
		expect map[string]any
	}

	var have_1 = WithsConfig{
		{
			Name:  "first_name",
			Value: "Mona",
		},
		{
			Name:  "middle_name",
			Value: "The",
		},
		{
			Name:  "last_name",
			Value: "Octocat",
		},
	}
	var expect_1 = map[string]any{
		"first_name":  "Mona",
		"middle_name": "The",
		"last_name":   "Octocat",
	}

	var have_2 = WithsConfig{}

	var tests = []Test{
		{"with defined working-directory", &have_1, expect_1},
		{"without empty working-directory", &have_2, nil},
		{"without undefined working-directory", nil, nil},
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
