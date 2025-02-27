// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestConcurrency(t *testing.T) {
	// type Test struct {
	// 	name   string
	// 	have   *ConcurrencyConfig
	// 	expect Concurrency
	// }

	// var group1 = "${{ github.workflow }}-${{ github.ref }}"
	// var group1Exp = hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal(group1),
	// 		},
	// 	},
	// }
	// var cancelInProgress1 = true
	// var cancelInProgress1Exp = hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.BoolVal(cancelInProgress1),
	// 		},
	// 	},
	// }
	// var group2 = "${{ github.workflow }}-${{ github.ref }}"
	// var group2Exp = hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal(group2),
	// 		},
	// 	},
	// }

	// var have1 = ConcurrencyConfig{
	// 	Group:            &group1Exp,
	// 	CancelInProgress: &cancelInProgress1Exp,
	// }
	// var expect1 = Concurrency{
	// 	Group:            &group1,
	// 	CancelInProgress: &cancelInProgress1,
	// }

	// var have2 = ConcurrencyConfig{
	// 	Group: &group2Exp,
	// }
	// var expect2 = Concurrency{
	// 	Group: &group2,
	// }

	// var tests = []Test{
	// 	{"with defined concurrency, group and Cancel-in-progress", &have1, expect1},
	// 	{"with defined concurrency and group", &have2, expect2},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err := tt.have.Parse()
	// 		if err != nil {
	// 			t.Fatal(err.Error())
	// 		}

	// 		if !reflect.DeepEqual(got, tt.expect) {
	// 			t.Fatal(tt.name)
	// 		}
	// 	})
	// }
}
