// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestOutputs(t *testing.T) {
	type Test struct {
		name   string
		have   OutputsConfig
		expect map[string]any
	}

	var output1 = OutputConfig{
		Name:  "name1",
		Value: cty.StringVal("val1"),
	}

	var output2 = OutputConfig{
		Name:  "name2",
		Value: cty.BoolVal(true),
	}

	var have1 = OutputsConfig{
		&output1,
	}

	var expect1 = map[string]any{
		"name1": "val1",
	}

	var have2 = OutputsConfig{
		&output1,
		&output2,
	}

	var expect2 = map[string]any{
		"name1": "val1",
		"name2": true,
	}

	var tests = []Test{
		{"with string output", have1, expect1},
		{"with multiple outputs", have2, expect2},
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
