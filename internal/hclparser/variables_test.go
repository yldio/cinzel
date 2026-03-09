// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestHCLVarsAddAndGet(t *testing.T) {
	hv := NewHCLVars()
	hv.Add("greeting", cty.StringVal("hello"))

	val, err := hv.GetValueByKey("greeting")
	if err != nil {
		t.Fatal(err)
	}

	if val.AsString() != "hello" {
		t.Fatalf("expected hello, got %s", val.AsString())
	}
}

func TestHCLVarsGetValueByKeyMissing(t *testing.T) {
	hv := NewHCLVars()
	_, err := hv.GetValueByKey("missing")

	if err == nil {
		t.Fatal("expected error for missing variable")
	}
}

func TestHCLVarsGetValueDispatch(t *testing.T) {
	hv := NewHCLVars()
	hv.Add("name", cty.StringVal("test"))

	t.Run("nil index returns by key", func(t *testing.T) {
		val, err := hv.GetValue("name", nil)
		if err != nil {
			t.Fatal(err)
		}

		if val.AsString() != "test" {
			t.Fatalf("expected test, got %s", val.AsString())
		}
	})

	t.Run("non-nil index delegates to GetValueByIndex", func(t *testing.T) {
		hv.Add("list", cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}))
		idx := int64(1)
		val, err := hv.GetValue("list", &idx)
		if err != nil {
			t.Fatal(err)
		}

		if val.AsString() != "b" {
			t.Fatalf("expected b, got %s", val.AsString())
		}
	})
}

func TestHCLVarsGetValueByIndex(t *testing.T) {
	hv := NewHCLVars()
	hv.Add("items", cty.ListVal([]cty.Value{cty.StringVal("first"), cty.StringVal("second")}))

	t.Run("valid index", func(t *testing.T) {
		val, err := hv.GetValueByIndex("items", 0)
		if err != nil {
			t.Fatal(err)
		}

		if val.AsString() != "first" {
			t.Fatalf("expected first, got %s", val.AsString())
		}
	})

	t.Run("out of range index", func(t *testing.T) {
		_, err := hv.GetValueByIndex("items", 5)

		if err == nil {
			t.Fatal("expected out of range error")
		}
	})

	t.Run("negative index", func(t *testing.T) {
		_, err := hv.GetValueByIndex("items", -1)

		if err == nil {
			t.Fatal("expected out of range error for negative index")
		}
	})

	t.Run("non-collection returns value directly", func(t *testing.T) {
		hv.Add("scalar", cty.StringVal("hello"))
		val, err := hv.GetValueByIndex("scalar", 0)
		if err != nil {
			t.Fatal(err)
		}

		if val.AsString() != "hello" {
			t.Fatalf("expected hello, got %s", val.AsString())
		}
	})

	t.Run("missing key", func(t *testing.T) {
		_, err := hv.GetValueByIndex("missing", 0)

		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})
}
