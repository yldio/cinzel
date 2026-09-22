// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
)

// expression parses src as a bare HCL expression.
func expression(t *testing.T, src string) hcl.Expression {
	t.Helper()

	expr, diags := hclsyntax.ParseExpression([]byte(src), "e.hcl", hcl.Pos{})

	if diags.HasErrors() {
		t.Fatalf("parsing %q: %v", src, diags)
	}

	return expr
}

// The name becomes a map key, so a value that is not a string used to reach
// cty.Value.AsString and panic with "not a string", taking the whole run down
// on a file an author could have been told about.
func TestANameThatIsNotAStringIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		src  string
		want string
	}{
		{name: "a number", src: "42", want: "number"},
		{name: "a bool", src: "true", want: "bool"},
		{name: "a list", src: `["a", "b"]`, want: "tuple"},
		{name: "an object", src: `{a = "b"}`, want: "object"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nameExpr := expression(t, tc.src)
			valueExpr := expression(t, `"v"`)

			envCfg := EnvListConfig{{Name: nameExpr, Value: valueExpr}}

			_, err := envCfg.Parse(hclparser.NewHCLVars())

			if err == nil {
				t.Fatal("expected an env block to be refused")
			}

			if !strings.Contains(err.Error(), "name must be a string") || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("env error does not say what was found: %v", err)
			}

			withCfg := WithListConfig{{Name: nameExpr, Value: valueExpr}}

			_, err = withCfg.Parse(hclparser.NewHCLVars())

			if err == nil {
				t.Fatal("expected a with block to be refused")
			}

			if !strings.Contains(err.Error(), "name must be a string") || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("with error does not say what was found: %v", err)
			}
		})
	}
}
