// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/zclconf/go-cty/cty"
)

// Every refusal here names something the author wrote in their own HCL, so the
// open-an-issue line cinzelerror.New adds by default sent them to the issue
// tracker over their own index or their own typo. This is what 20f8a92 fixed
// for .cinzelrc.yaml, and errors.go had no marking at all.
func TestExpressionErrorsAreTheAuthorsToFix(t *testing.T) {
	for _, tc := range []struct{ name, expr string }{
		{"an index that is not a number", `variable.envs["prod"]`},
		{"an index on something with no positions", "variable.envs[0]"},
		{"a reference into a value", "variable.envs.prod"},
		{"a variable nobody declared", "variable.nope"},
		{"an index past the end", "variable.list[9]"},
		{"a division by zero", "5 / variable.zero"},
		{"an operator cinzel does not evaluate", "5 % 2"},
		{"an operand that is not a number", `variable.list / 2`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseExprErr(t, tc.expr)

			if !cinzelerror.IsUserInput(err) {
				t.Errorf("error = %v, which asks the author to open an issue about their own HCL", err)
			}
		})
	}
}

// The mark has to discriminate, not blanket the package: an expression cinzel
// has no case for is cinzel's to answer for, and that one keeps the line.
func TestAnExpressionCinzelCannotReadIsNotTheAuthorsFault(t *testing.T) {
	err := parseExprErr(t, `"a" * 2`)

	if cinzelerror.IsUserInput(err) {
		t.Errorf("error = %v, but an expression cinzel has no case for is cinzel's", err)
	}
}

// hclsyntax builds a TemplateWrapExpr, not a TemplateExpr, when a template
// holds one interpolation and nothing else. There was no case for it, so
// "${variable.list[0]}" came back "missing hcl type found" with a raw pointer
// dump after it, where the same reference written bare resolved.
func TestASoleInterpolationResolves(t *testing.T) {
	hp := parseExpr(t, "${variable.list[0]}", true)

	if err := hp.Parse(); err != nil {
		t.Fatalf("Parse() error = %v, want the wrapped value", err)
	}

	if got := hp.Result(); got.AsString() != "a" {
		t.Errorf("Result() = %#v, want the element the bare reference gives", got)
	}
}

// parseExpr builds a parser over expr against a fixed set of variables.
// A quoted expr is written as a string template, which is what produces the
// wrap the case above covers.
func parseExpr(t *testing.T, expr string, quoted bool) *HCLParser {
	t.Helper()

	src := "value = " + expr + "\n"
	if quoted {
		src = `value = "` + expr + "\"\n"
	}

	file, diags := hclsyntax.ParseConfig([]byte(src), "test.hcl", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		t.Fatalf("body is %T", file.Body)
	}

	hv := NewHCLVars()
	hv.Add("envs", cty.ObjectVal(map[string]cty.Value{"prod": cty.StringVal("x")}))
	hv.Add("list", cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}))
	hv.Add("zero", cty.NumberIntVal(0))

	return New(body.Attributes["value"].Expr, hv)
}

// parseExprErr requires expr to be refused and returns the refusal.
func parseExprErr(t *testing.T, expr string) error {
	t.Helper()

	err := parseExpr(t, expr, false).Parse()
	if err == nil {
		t.Fatalf("Parse() error = nil, want %q refused", expr)
	}

	return err
}
