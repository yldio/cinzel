// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestCredentials(t *testing.T) {
	type Test struct {
		name   string
		have   CredentialsConfig
		expect *Credentials
	}

	username := hclsyntax.TemplateExpr{
		Parts: []hclsyntax.Expression{
			&hclsyntax.LiteralValueExpr{
				Val: cty.StringVal("${{ github.actor }}"),
			},
		},
	}

	password := hclsyntax.TemplateExpr{
		Parts: []hclsyntax.Expression{
			&hclsyntax.LiteralValueExpr{
				Val: cty.StringVal("${{ secrets.github_token }}"),
			},
		},
	}

	have1 := CredentialsConfig{
		Username: &username,
		Password: &password,
	}

	expect1 := &Credentials{
		Username: "${{ github.actor }}",
		Password: "${{ secrets.github_token }}",
	}

	var tests = []Test{
		{"with defined credentials", have1, expect1},
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
