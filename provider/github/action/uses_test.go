// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func TestUses(t *testing.T) {
	t.Run("test 1", func(t *testing.T) {
		action := []byte(`"my-action"`)
		version := []byte(`"v1.1.1"`)
		expect := "my-action@v1.1.1"

		actionExpr, diags := hclsyntax.ParseExpression(action, "action.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		versionExpr, diags := hclsyntax.ParseExpression(version, "version.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		config := UsesListConfig{
			{
				Action:  actionExpr,
				Version: versionExpr,
			},
		}

		hv := hclparser.NewHCLVars()

		val, err := config.Parse(hv)
		if err != nil {
			t.FailNow()
		}

		if val.AsString() != expect {
			t.FailNow()
		}
	})

	t.Run("local action without version", func(t *testing.T) {
		action := []byte(`"./.github/actions/my-action"`)

		actionExpr, diags := hclsyntax.ParseExpression(action, "action.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		config := UsesListConfig{
			{
				Action: actionExpr,
			},
		}

		hv := hclparser.NewHCLVars()

		val, err := config.Parse(hv)
		if err != nil {
			t.Fatal(err)
		}

		if val.AsString() != "./.github/actions/my-action" {
			t.Fatalf("expected './.github/actions/my-action', got %q", val.AsString())
		}
	})

	t.Run("test 3", func(t *testing.T) {
		version := []byte(`"v1.1.1"`)
		expected := "action must be set"

		versionExpr, diags := hclsyntax.ParseExpression(version, "version.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		config := UsesListConfig{
			{
				Version: versionExpr,
			},
		}

		hv := hclparser.NewHCLVars()

		val, err := config.Parse(hv)

		if err.Error() != expected {
			t.FailNow()
		}

		if val != cty.NilVal {
			t.FailNow()
		}
	})

	t.Run("test 4", func(t *testing.T) {
		action := []byte(`"my-action"`)
		version := []byte(`1`)
		expected := "unsupported type, expected string, found number"

		actionExpr, diags := hclsyntax.ParseExpression(action, "action.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		versionExpr, diags := hclsyntax.ParseExpression(version, "version.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		config := UsesListConfig{
			{
				Action:  actionExpr,
				Version: versionExpr,
			},
		}

		hv := hclparser.NewHCLVars()

		val, err := config.Parse(hv)

		if err.Error() != expected {
			t.FailNow()
		}

		if val != cty.NilVal {
			t.FailNow()
		}
	})

	t.Run("test 5", func(t *testing.T) {
		action := []byte(`1`)
		version := []byte(`"v1.1.1"`)
		expected := "unsupported type, expected string, found number"

		actionExpr, diags := hclsyntax.ParseExpression(action, "action.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		versionExpr, diags := hclsyntax.ParseExpression(version, "version.hcl", hcl.Pos{})

		if diags.HasErrors() {
			t.FailNow()
		}

		config := UsesListConfig{
			{
				Action:  actionExpr,
				Version: versionExpr,
			},
		}

		hv := hclparser.NewHCLVars()

		val, err := config.Parse(hv)

		if err.Error() != expected {
			t.FailNow()
		}

		if val != cty.NilVal {
			t.FailNow()
		}
	})
}
