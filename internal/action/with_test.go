// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestWith(t *testing.T) {
	// type Test struct {
	// 	name   string
	// 	have   *WithsConfig
	// 	expect map[string]any
	// }

	// var name1 any = "first_name"
	// var value1 any = "Mona"
	// var name2 any = "middle_name"
	// var value2 any = "The"
	// var name3 any = "last_name"
	// var value3 any = "Octocat"

	// var firstName = value1.(string)
	// var middleName = value2.(string)
	// var lastName = value3.(string)

	// var have1 = WithsConfig{
	// 	{
	// 		Name: &hclsyntax.TemplateExpr{
	// 			Parts: []hclsyntax.Expression{
	// 				&hclsyntax.LiteralValueExpr{
	// 					Val: cty.StringVal(name1.(string)),
	// 				},
	// 			},
	// 		},
	// 		Value: &hclsyntax.TemplateExpr{
	// 			Parts: []hclsyntax.Expression{
	// 				&hclsyntax.LiteralValueExpr{
	// 					Val: cty.StringVal(value1.(string)),
	// 				},
	// 			},
	// 		},
	// 	},
	// 	{
	// 		Name: &hclsyntax.TemplateExpr{
	// 			Parts: []hclsyntax.Expression{
	// 				&hclsyntax.LiteralValueExpr{
	// 					Val: cty.StringVal(name2.(string)),
	// 				},
	// 			},
	// 		},
	// 		Value: &hclsyntax.TemplateExpr{
	// 			Parts: []hclsyntax.Expression{
	// 				&hclsyntax.LiteralValueExpr{
	// 					Val: cty.StringVal(value2.(string)),
	// 				},
	// 			},
	// 		},
	// 	},
	// 	{
	// 		Name: &hclsyntax.TemplateExpr{
	// 			Parts: []hclsyntax.Expression{
	// 				&hclsyntax.LiteralValueExpr{
	// 					Val: cty.StringVal(name3.(string)),
	// 				},
	// 			},
	// 		},
	// 		Value: &hclsyntax.TemplateExpr{
	// 			Parts: []hclsyntax.Expression{
	// 				&hclsyntax.LiteralValueExpr{
	// 					Val: cty.StringVal(value3.(string)),
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// var expect1 = map[string]any{
	// 	"first_name":  &firstName,
	// 	"middle_name": &middleName,
	// 	"last_name":   &lastName,
	// }

	// var have2 = WithsConfig{}

	// var tests = []Test{
	// 	{"with defined with", &have1, expect1},
	// 	{"without empty with", &have2, nil},
	// 	{"without undefined with", nil, nil},
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
