// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"math/big"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func TestStepEnvSuccess(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect cty.Value
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`{
			name = "my-name"
			value = "my-value"
			}`),
			cty.ObjectVal(map[string]cty.Value{
				"name":  cty.StringVal("my-name"),
				"value": cty.StringVal("my-value"),
			}),
		},
		{
			"test 2",
			[]byte(`{
			name = "my-name"
			value = 11
			}`),
			cty.ObjectVal(map[string]cty.Value{
				"name":  cty.StringVal("my-name"),
				"value": cty.NumberVal(big.NewFloat(11)),
			}),
		},
		{
			"test 3",
			[]byte(`{
			name = "my-name"
			value = true
			}`),
			cty.ObjectVal(map[string]cty.Value{
				"name":  cty.StringVal("my-name"),
				"value": cty.BoolVal(true),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression(tt.have, "example.hcl", hcl.Pos{})

			if diags.HasErrors() {
				t.FailNow()
			}

			hv := hclparser.NewHCLVars()
			hv.Add("test", cty.ObjectVal(map[string]cty.Value{
				"name":  cty.StringVal("my-name"),
				"value": cty.NumberVal(big.NewFloat(11)),
			}))

			stepParsed := Step{}
			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

			if err := stepParsed.parseEnv(hp.Result()); err != nil {
				t.FailNow()
			}

			if !stepParsed.Env.RawEquals(tt.expect) {
				t.FailNow()
			}
		})
	}
}

func TestStepEnvFailure(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect cty.Value
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`1`),
			cty.NilVal,
		},
		{
			"test 2",
			[]byte(`"my-name"`),
			cty.NilVal,
		},
		{
			"test 3",
			[]byte(`true`),
			cty.NilVal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression(tt.have, "example.hcl", hcl.Pos{})

			if diags.HasErrors() {
				t.FailNow()
			}

			hv := hclparser.NewHCLVars()
			hv.Add("test", cty.StringVal("my-id"))

			stepParsed := Step{}
			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

			if err := stepParsed.parseEnv(hp.Result()); err == nil {
				t.FailNow()
			}

			if !stepParsed.Env.RawEquals(tt.expect) {
				t.FailNow()
			}
		})
	}
}
