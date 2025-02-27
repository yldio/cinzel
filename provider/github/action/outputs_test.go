// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

// import (
// 	"reflect"
// 	"testing"

// 	"github.com/hashicorp/hcl/v2/hclsyntax"
// 	"github.com/zclconf/go-cty/cty"
// )

// func TestOutputs(t *testing.T) {
// 	type Test struct {
// 		name   string
// 		have   *OutputsConfig
// 		expect map[string]any
// 	}

// 	var value1 any = "val1"
// 	var value2 any = true

// 	var output1 = OutputConfig{
// 		Name: "name1",
// 		Value: &hclsyntax.TemplateExpr{
// 			Parts: []hclsyntax.Expression{
// 				&hclsyntax.LiteralValueExpr{
// 					Val: cty.StringVal(value1.(string)),
// 				},
// 			},
// 		},
// 	}

// 	var output2 = OutputConfig{
// 		Name: "name2",
// 		Value: &hclsyntax.TemplateExpr{
// 			Parts: []hclsyntax.Expression{
// 				&hclsyntax.LiteralValueExpr{
// 					Val: cty.BoolVal(value2.(bool)),
// 				},
// 			},
// 		},
// 	}

// 	var have1 = OutputsConfig{
// 		&output1,
// 	}

// 	var expect1 = map[string]any{
// 		"name1": &value1,
// 	}

// 	var have2 = OutputsConfig{
// 		&output1,
// 		&output2,
// 	}

// 	var expect2 = map[string]any{
// 		"name1": &value1,
// 		"name2": &value2,
// 	}

// 	var tests = []Test{
// 		{"with string output", &have1, expect1},
// 		{"with multiple outputs", &have2, expect2},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.have.Parse()
// 			if err != nil {
// 				t.Fatal(err.Error())
// 			}

// 			if !reflect.DeepEqual(got, tt.expect) {
// 				t.Fatal(tt.name)
// 			}
// 		})
// 	}
// }
