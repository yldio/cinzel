// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestEnvironment(t *testing.T) {
	// type Test struct {
	// 	name   string
	// 	have   *EnvironmentConfig
	// 	expect any
	// }

	// var name1 = "staging_environment"
	// var name2 = "production_environment"
	// var url2 = "${{ steps.step_id.outputs.url_output }}"

	// var nameExpression1 = hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal(name1),
	// 		},
	// 	},
	// }

	// var nameExpression2 = hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal(name2),
	// 		},
	// 	},
	// }

	// var urlExpression2 = hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal(url2),
	// 		},
	// 	},
	// }

	// var have1 = EnvironmentConfig{
	// 	Name: &nameExpression1,
	// }
	// var expect1 = name1

	// var have2 = EnvironmentConfig{
	// 	Name: &nameExpression2,
	// 	Url:  &urlExpression2,
	// }
	// var expect2 = map[string]string{
	// 	"name": "production_environment",
	// 	"url":  "${{ steps.step_id.outputs.url_output }}",
	// }

	// var tests = []Test{
	// 	{"with defined environment", &have1, &expect1},
	// 	{"without empty environment", &have2, &expect2},
	// 	{"without undefined environment", nil, nil},
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
