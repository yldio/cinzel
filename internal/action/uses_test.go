// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestUses(t *testing.T) {
	// type Test struct {
	// 	name   string
	// 	have   *UsesConfig
	// 	expect *string
	// }

	// var action = "actions/checkout"
	// var version = "v4"

	// var have1 = UsesConfig{
	// 	Action: &hclsyntax.TemplateExpr{
	// 		Parts: []hclsyntax.Expression{
	// 			&hclsyntax.LiteralValueExpr{
	// 				Val: cty.StringVal(action),
	// 			},
	// 		},
	// 	},
	// 	Version: &hclsyntax.TemplateExpr{
	// 		Parts: []hclsyntax.Expression{
	// 			&hclsyntax.LiteralValueExpr{
	// 				Val: cty.StringVal(version),
	// 			},
	// 		},
	// 	},
	// }
	// var expect1 = fmt.Sprintf("%s@%s", action, version)

	// var tests = []Test{
	// 	{"with defined uses", &have1, &expect1},
	// 	{"without undefined uses", nil, nil},
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
