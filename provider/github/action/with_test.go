// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"math/big"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func TestWith(t *testing.T) {
	type Test struct {
		name        string
		haveKey     []byte
		haveValue   []byte
		expectKey   cty.Value
		expectValue cty.Value
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`"my-name"`),
			[]byte(`"my-value"`),
			cty.StringVal("my-name"),
			cty.StringVal("my-value"),
		},
		{
			"test 2",
			[]byte(`"my-name"`),
			[]byte(`true`),
			cty.StringVal("my-name"),
			cty.BoolVal(true),
		},
		{
			"test 3",
			[]byte(`"my-name"`),
			[]byte(`11`),
			cty.StringVal("my-name"),
			cty.NumberVal(big.NewFloat(11)),
		},
		{
			"test 4",
			[]byte(`"my-name"`),
			[]byte(`22.22`),
			cty.StringVal("my-name"),
			cty.NumberVal(big.NewFloat(22.22)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyExpr, diags := hclsyntax.ParseExpression(tt.haveKey, "key.hcl", hcl.Pos{})

			if diags.HasErrors() {
				t.FailNow()
			}

			valueExpr, diags := hclsyntax.ParseExpression(tt.haveValue, "value.hcl", hcl.Pos{})

			if diags.HasErrors() {
				t.FailNow()
			}

			config := WithListConfig{
				{
					Name:  keyExpr,
					Value: valueExpr,
				},
			}

			hv := hclparser.NewHCLVars()

			val, err := config.Parse(hv)
			if err != nil {
				t.FailNow()
			}

			iter := val.ElementIterator()

			for iter.Next() {
				key, value := iter.Element()

				if !key.RawEquals(tt.expectKey) {
					t.FailNow()
				}

				if !value.RawEquals(tt.expectValue) {
					t.FailNow()
				}
			}
		})
	}
}
