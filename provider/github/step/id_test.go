// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/internal/yamlwriter"
	"github.com/zclconf/go-cty/cty"
)

func TestStepIdSuccess(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect cty.Value
		yaml   string
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`"my-id"`),
			cty.StringVal("my-id"),
			"id: my-id\n",
		},
		{
			"test 2",
			[]byte(`"my_id"`),
			cty.StringVal("my_id"),
			"id: my_id\n",
		},
		{
			"test 3",
			[]byte(`"_my-id"`),
			cty.StringVal("_my-id"),
			"id: _my-id\n",
		},
		{
			"test 4",
			[]byte(`var.test`),
			cty.StringVal("my-id"),
			"id: my-id\n",
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

			if err := stepParsed.parseId(hp.Result()); err != nil {
				t.FailNow()
			}

			if stepParsed.Id != tt.expect {
				t.FailNow()
			}

			out, err := yamlwriter.Marshal(stepParsed)
			if err != nil {
				t.FailNow()
			}

			if string(out) != tt.yaml {
				t.FailNow()
			}
		})
	}
}

func TestStepIdFailure(t *testing.T) {
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
			[]byte(`"-my-id"`),
			cty.NilVal,
		},
		{
			"test 3",
			[]byte(`"123my-id"`),
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
			hv.Add("test", cty.BoolVal(true))

			stepParsed := Step{}
			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

			if err := stepParsed.parseId(hp.Result()); err == nil {
				t.FailNow()
			}

			if stepParsed.Id != tt.expect {
				t.FailNow()
			}
		})
	}
}
