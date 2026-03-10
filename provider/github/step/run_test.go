// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/internal/yamlwriter"
	"github.com/zclconf/go-cty/cty"
)

func TestStepRunSuccess(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect cty.Value
		yaml   string
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`"npm install"`),
			cty.StringVal("npm install"),
			"run: npm install\n",
		},
		{
			"test 2",
			[]byte(`var.test`),
			cty.StringVal("npm install"),
			"run: npm install\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression(tt.have, "example.hcl", hcl.Pos{})

			if diags.HasErrors() {
				t.FailNow()
			}

			hv := hclparser.NewHCLVars()
			hv.Add("test", cty.StringVal("npm install"))

			stepParsed := Step{}
			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

			if err := stepParsed.parseRun(hp.Result()); err != nil {
				t.FailNow()
			}

			if stepParsed.Run != tt.expect {
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

func TestStepRunFailure(t *testing.T) {
	type Test struct {
		name   string
		have   []byte
		expect cty.Value
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`true`),
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
			hv.Add("test", cty.BoolVal(true))

			stepParsed := Step{}
			hp := hclparser.New(expr, hv)

			if err := hp.Parse(); err != nil {
				t.FailNow()
			}

			if err := stepParsed.parseRun(hp.Result()); err == nil {
				t.FailNow()
			}

			if stepParsed.Run != tt.expect {
				t.FailNow()
			}
		})
	}
}
