// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestVariableConfigParse(t *testing.T) {
	t.Run("registers variables from config", func(t *testing.T) {
		config := &VariablesConfig{
			{Id: "key1", Value: cty.StringVal("value1")},
			{Id: "key2", Value: cty.BoolVal(true)},
			{Id: "key3", Value: cty.NumberIntVal(42)},
		}

		hv := NewHCLVars()

		if err := config.Parse(hv); err != nil {
			t.Fatal(err)
		}

		val, err := hv.GetValueByKey("key1")
		if err != nil {
			t.Fatal(err)
		}

		if val.AsString() != "value1" {
			t.Fatalf("expected value1, got %s", val.AsString())
		}

		val, err = hv.GetValueByKey("key2")
		if err != nil {
			t.Fatal(err)
		}

		if val.True() != true {
			t.Fatal("expected true")
		}

		val, err = hv.GetValueByKey("key3")
		if err != nil {
			t.Fatal(err)
		}
		i, _ := val.AsBigFloat().Int64()

		if i != 42 {
			t.Fatalf("expected 42, got %d", i)
		}
	})

	t.Run("nil config is a no-op", func(t *testing.T) {
		var config *VariablesConfig
		hv := NewHCLVars()

		if err := config.Parse(hv); err != nil {
			t.Fatalf("expected nil config to be a no-op, got %v", err)
		}
	})

	t.Run("empty config is a no-op", func(t *testing.T) {
		config := &VariablesConfig{}
		hv := NewHCLVars()

		if err := config.Parse(hv); err != nil {
			t.Fatalf("expected empty config to be a no-op, got %v", err)
		}
	})
}
