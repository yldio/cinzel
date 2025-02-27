// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package variable

import (
	"testing"

	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

func TestVariables(t *testing.T) {
	type Test struct {
		name   string
		have   *VariablesConfig
		expect *variables.Variables
	}

	var variable1 = cty.StringVal("value1")
	var variable2 = cty.BoolVal(true)
	var variable3 = cty.TupleVal([]cty.Value{cty.StringVal("value2"), cty.StringVal("value3")})

	var have1 = VariablesConfig{
		{
			Id:    "key1",
			Value: variable1,
		},
		{
			Id:    "key2",
			Value: variable2,
		},
		{
			Id:    "key3",
			Value: variable3,
		},
	}
	var expect1 = variables.Instance()

	var tests = []Test{
		{"with defined Id", &have1, expect1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// got := tt.have.Parse()

			// if !reflect.DeepEqual(got, tt.expect) {
			// 	// t.Fatal(tt.name)
			// }
		})
	}
}
