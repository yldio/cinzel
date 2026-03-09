// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestParsers(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect any
	}

	var tests = []Test{
		{"[BinaryOpExpr] parse an OpAdd", []byte(`1 + 2`), cty.NumberIntVal(3)},
		{"[BinaryOpExpr] parse an OpAdd 2", []byte(`1 + (2 + 1)`), cty.NumberIntVal(4)},
		{"[BinaryOpExpr] parse unary negation as operand", []byte(`-5 + 3`), cty.NumberIntVal(-2)},
		{"[BinaryOpExpr] parse an OpSubtract", []byte(`2 - 1`), cty.NumberIntVal(1)},
		{"[BinaryOpExpr] parse an OpMultiply", []byte(`2 * 2`), cty.NumberIntVal(4)},
		{"[BinaryOpExpr] parse an OpMultiply", []byte(`2 * 2`), cty.NumberIntVal(4)},
		{"[BinaryOpExpr] parse an OpDivide", []byte(`4 / 2`), cty.NumberFloatVal(2)},
		{"[BinaryOpExpr] parse an OpEqual", []byte(`2 == 2`), cty.BoolVal(true)},
		{"[BinaryOpExpr] parse an OpNotEqual", []byte(`4 != 2`), cty.BoolVal(true)},
		{"[BinaryOpExpr] parse an OpGreaterThan", []byte(`4 > 2`), cty.BoolVal(true)},
		{"[BinaryOpExpr] parse an OpGreaterThanOrEqual", []byte(`2 >= 2`), cty.BoolVal(true)},
		{"[BinaryOpExpr] parse an OpLessThan", []byte(`2 < 4`), cty.BoolVal(true)},
		{"[BinaryOpExpr] parse an OpLessThanOrEqual", []byte(`2 <= 2`), cty.BoolVal(true)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			exp, diags := hclsyntax.ParseExpression(test.have, "", hcl.Pos{})

			if diags.HasErrors() {
				t.Fatal(diags.Error())
			}

			hv := NewHCLVars()

			hp := New(exp, hv)

			if err := hp.Parse(); err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(hp.Result(), test.expect) {
				t.Fatalf("expected %s but got %s", test.expect, hp.Result())
			}
		})
	}
}

func TestBinaryOpExprWithVariables(t *testing.T) {
	tests := []struct {
		name   string
		expr   []byte
		vars   map[string]cty.Value
		expect cty.Value
	}{
		{
			name:   "variable reference as LHS operand",
			expr:   []byte(`var.timeout + 5`),
			vars:   map[string]cty.Value{"timeout": cty.NumberIntVal(10)},
			expect: cty.NumberIntVal(15),
		},
		{
			name:   "variable reference as RHS operand",
			expr:   []byte(`20 - var.offset`),
			vars:   map[string]cty.Value{"offset": cty.NumberIntVal(3)},
			expect: cty.NumberIntVal(17),
		},
		{
			name:   "variable reference on both sides",
			expr:   []byte(`var.a * var.b`),
			vars:   map[string]cty.Value{"a": cty.NumberIntVal(4), "b": cty.NumberIntVal(3)},
			expect: cty.NumberIntVal(12),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, diags := hclsyntax.ParseExpression(tt.expr, "", hcl.Pos{})

			if diags.HasErrors() {
				t.Fatal(diags.Error())
			}

			hv := NewHCLVars()

			for k, v := range tt.vars {
				hv.Add(k, v)
			}

			hp := New(exp, hv)

			if err := hp.Parse(); err != nil {
				t.Fatal(err)
			}

			if !hp.Result().RawEquals(tt.expect) {
				t.Fatalf("expected %s but got %s", tt.expect.GoString(), hp.Result().GoString())
			}
		})
	}
}

func TestParserHandlesTupleAndUnaryExpressions(t *testing.T) {
	tests := []struct {
		name   string
		expr   []byte
		expect cty.Value
	}{
		{
			name:   "tuple expression",
			expr:   []byte(`["linux", "darwin"]`),
			expect: cty.TupleVal([]cty.Value{cty.StringVal("linux"), cty.StringVal("darwin")}),
		},
		{
			name:   "unary expression",
			expr:   []byte(`-10`),
			expect: cty.NumberIntVal(-10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, diags := hclsyntax.ParseExpression(tt.expr, "", hcl.Pos{})

			if diags.HasErrors() {
				t.Fatal(diags.Error())
			}

			hv := NewHCLVars()
			hp := New(exp, hv)

			if err := hp.Parse(); err != nil {
				t.Fatal(err)
			}

			if !hp.Result().RawEquals(tt.expect) {
				t.Fatalf("expected %s but got %s", tt.expect.GoString(), hp.Result().GoString())
			}
		})
	}
}
