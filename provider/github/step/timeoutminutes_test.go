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

func TestStepTimeoutMinutesSuccess(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect cty.Value
		yaml   string
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`360`),
			cty.NumberUIntVal(360),
			"timeout-minutes: 360\n",
		},
		{
			"test 2",
			[]byte(`var.test`),
			cty.NumberUIntVal(360),
			"timeout-minutes: 360\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression(tt.have, "example.hcl", hcl.Pos{})
			if diags.HasErrors() {
				t.FailNow()
			}

			hv := hclparser.NewHCLVars()
			hv.Add("test", cty.NumberUIntVal(360))

			stepParsed := Step{}
			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

			if err := stepParsed.parseTimeoutMinutes(hp.Result()); err != nil {
				t.FailNow()
			}

			if stepParsed.TimeoutMinutes.AsBigFloat().Cmp(tt.expect.AsBigFloat()) != 0 {
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

func TestStepTimeoutMinutesFailure(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect cty.Value
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`"one"`),
			cty.NilVal,
		},
		{
			"test 2",
			[]byte(`var.test`),
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
			hv.Add("test", cty.StringVal("one"))

			stepParsed := Step{}
			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

			if err := stepParsed.parseTimeoutMinutes(hp.Result()); err == nil {
				t.FailNow()
			}

			if stepParsed.TimeoutMinutes != tt.expect {
				t.FailNow()
			}
		})
	}
}
