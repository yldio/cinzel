// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDefaults(t *testing.T) {
	type Test struct {
		name   string
		have   *DefaultsConfig
		expect *Defaults
	}
	shell := "bash"
	workingDirectory := "./scripts"

	shellExpression := hclsyntax.TemplateExpr{
		Parts: []hclsyntax.Expression{
			&hclsyntax.LiteralValueExpr{
				Val: cty.StringVal(shell),
			},
		},
	}

	workingDirectoryExpression := hclsyntax.TemplateExpr{
		Parts: []hclsyntax.Expression{
			&hclsyntax.LiteralValueExpr{
				Val: cty.StringVal(workingDirectory),
			},
		},
	}

	var have1 = DefaultsConfig{
		Run: &DefaultsRunConfig{
			Shell:            &shellExpression,
			WorkingDirectory: &workingDirectoryExpression,
		},
	}
	var expect1 = Defaults{
		Run: &Run{
			Shell:            &shell,
			WorkingDirectory: &workingDirectory,
		},
	}

	var have2 = DefaultsConfig{
		Run: &DefaultsRunConfig{
			Shell: &shellExpression,
		},
	}
	var expect2 = Defaults{
		Run: &Run{
			Shell: &shell,
		},
	}

	var have3 = DefaultsConfig{
		Run: &DefaultsRunConfig{
			WorkingDirectory: &workingDirectoryExpression,
		},
	}
	var expect3 = Defaults{
		Run: &Run{
			WorkingDirectory: &workingDirectory,
		},
	}

	var tests = []Test{
		{"with defined defaults, run shell and working-directory", &have1, &expect1},
		{"with defined defaults, run shell", &have2, &expect2},
		{"with defined defaults and run working-directory", &have3, &expect3},
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
