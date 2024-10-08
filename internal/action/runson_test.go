// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"reflect"
	"testing"
)

func TestRunsOn(t *testing.T) {
	type Test struct {
		name   string
		have   *RunsOnConfig
		expect any
	}

	// var onUbuntu any = "ubuntu-latest"
	// var onUbuntuStr string = onUbuntu.(string)
	// var onLinux any = "linux"
	// var onLinuxStr string = onLinux.(string)
	// var onGroup1 any = "ubuntu-runners"
	// var onLabels1 any = "ubuntu-20.04-16core"

	// var have1 = RunsOnConfig{
	// 	On: &hclsyntax.TemplateExpr{
	// 		Parts: []hclsyntax.Expression{
	// 			&hclsyntax.LiteralValueExpr{
	// 				Val: cty.StringVal(onUbuntuStr),
	// 			},
	// 		},
	// 	},
	// }
	// var expect1 = &onUbuntu

	// var have2 = RunsOnConfig{
	// 	On: &hclsyntax.TemplateExpr{
	// 		Parts: []hclsyntax.Expression{
	// 			&hclsyntax.LiteralValueExpr{
	// 				Val: cty.TupleVal([]cty.Value{cty.StringVal(onUbuntuStr), cty.StringVal(onLinuxStr)}),
	// 			},
	// 		},
	// 	},
	// }

	// var expect2 any = []any{&onUbuntu, &onLinux}

	// var have3 = RunsOnConfig{
	// 	OnGroup: &hclsyntax.TemplateExpr{
	// 		Parts: []hclsyntax.Expression{
	// 			&hclsyntax.LiteralValueExpr{
	// 				Val: cty.StringVal(on3.(string)),
	// 			},
	// 		},
	// 	},
	// }
	// var expect3 = map[string]any{
	// 	"group": "ubuntu-runners",
	// }

	var tests = []Test{
		// {"with defined runs-on as string", &have1, expect1},
		// {"with defined runs-on as an array of strings", &have2, expect2},
		// {"with runs-on group", &have3, expect3},
		// {"with runs-on labels", &have3, expect3},
		// {"with runs-on group and labels", &have3, expect3},
		// {"without runs-on", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatal(tt.name)
			}
		})
	}
}
