// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func HelperGetInt64Ref(idx int64) *int64 {
	index := new(int64)
	*index = idx
	return index
}

func TestActoParsers(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect any
	}

	var tests = []Test{
		{"parse as positive number", []byte(`10`), uint64(10)},
		{"parse as negative number", []byte(`-10`), int64(-10)},
		{"parse as floating positive number", []byte(`10.5`), float64(10.5)},
		{"parse as floating negative number", []byte(`-12.5`), float64(-12.5)},
		{"parse as bool", []byte(`true`), true},
		{"parse as string", []byte(`"my-value"`), "my-value"},
		{"parse as list of strings", []byte(`["value-1", "value-2"]`), []string{"value-1", "value-2"}},
		{"parse as list of positive numbers", []byte(`[10, 12]`), []uint64{uint64(10), uint64(12)}},
		{"parse as list of negative numbers", []byte(`[-10, -12]`), []int64{int64(-10), int64(-12)}},
		{"parse as list of floating numbers", []byte(`[10.5, -12.5]`), []float64{float64(10.5), float64(-12.5)}},
		{"parse as list of floating numbers", []byte(`[true, false]`), []bool{true, false}},
		{"parse as variable", []byte(`variable.list_of_os`), ActoVariableRef{"variable", "list_of_os", nil}},
		{"parse as variable element", []byte(`variable.list_of_os[0]`), ActoVariableRef{"variable", "list_of_os", HelperGetInt64Ref(0)}},
		{"parse as list of variable elements", []byte(`[variable.list_of_os[0], variable.list_of_os[1]]`), []ActoVariableRef{{"variable", "list_of_os", HelperGetInt64Ref(0)}, {"variable", "list_of_os", HelperGetInt64Ref(1)}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			exp, diags := hclsyntax.ParseExpression(test.have, "", hcl.Pos{})
			if diags.HasErrors() {
				t.Fatal(diags.Error())
			}

			acto := NewActo(exp)

			if err := acto.Parse(); err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(acto.Result, test.expect) {
				t.Fatalf("expected %s but got %s", test.expect, acto.Result)
			}
		})
	}
}
