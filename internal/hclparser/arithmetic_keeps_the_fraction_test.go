// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
)

// Add, subtract and multiply read each side with big.Float.Int64, which
// truncates, so the operands were whole before the operation ever happened:
// "1.5 + 2.5" was computed as "1 + 2" and came back as 3, and "0.5 + 0.5" came
// back as 0, which the workflow validator then refused as "must be greater
// than zero" over arithmetic the author had written correctly. Divide never had
// it, because it reads both sides as Float64.
func TestArithmeticKeepsTheFraction(t *testing.T) {
	for _, tc := range []struct {
		expr string
		want string
	}{
		{expr: "1.5 + 2.5", want: "4"},
		{expr: "0.5 + 0.5", want: "1"},
		{expr: "2.9 + 0", want: "2.9"},
		{expr: "0.4 * 10", want: "4"},
		{expr: "3.7 - 0.7", want: "3"},
		{expr: "1.5 * 1.5", want: "2.25"},

		// The whole cases have to keep writing as integers: a minute count
		// going out as "6.0" would change every golden that holds one.
		{expr: "2 * 3", want: "6"},
		{expr: "10 - 4", want: "6"},
		{expr: "1000 + 1", want: "1001"},

		// Divide was already right and has to stay that way.
		{expr: "7 / 2", want: "3.5"},
	} {
		t.Run(tc.expr, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression([]byte(tc.expr), "t.hcl", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				t.Fatalf("ParseExpression() error = %v", diags)
			}

			binary, ok := expr.(*hclsyntax.BinaryOpExpr)
			if !ok {
				t.Fatalf("want a binary op, got %T", expr)
			}

			got, err := hclparser.NewBinaryOpExpr(binary, hclparser.NewHCLVars()).Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if s := got.AsBigFloat().Text('g', -1); s != tc.want {
				t.Errorf("%s = %s, want %s", tc.expr, s, tc.want)
			}
		})
	}
}
