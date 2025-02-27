// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestEnv(t *testing.T) {
	// type Test struct {
	// 	name   string
	// 	have   *EnvConfig
	// 	expect Env
	// }

	// var have1 = EnvConfig{
	// 	Variable: []VariableConfig{
	// 		{
	// 			Name: "NODE_ENV",
	// 			Value: &hclsyntax.TemplateExpr{
	// 				Parts: []hclsyntax.Expression{
	// 					&hclsyntax.LiteralValueExpr{
	// 						Val: cty.StringVal("development"),
	// 					},
	// 				},
	// 			},
	// 		},
	// 		{
	// 			Name: "TOKEN",
	// 			Value: &hclsyntax.TemplateExpr{
	// 				Parts: []hclsyntax.Expression{
	// 					&hclsyntax.LiteralValueExpr{
	// 						Val: cty.StringVal("${{ secrets.token }}"),
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// var expect1 = Env{
	// 	"NODE_ENV": "development",
	// 	"TOKEN":    "${{ secrets.token }}",
	// }

	// var tests = []Test{
	// 	{"with defined env", &have1, expect1},
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
