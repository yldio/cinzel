// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestParseCtyValueString(t *testing.T) {
	val, err := ParseCtyValue(cty.StringVal("hello"), []string{"string"})
	if err != nil {
		t.Fatal(err)
	}

	if val != "hello" {
		t.Fatalf("expected hello, got %v", val)
	}
}

func TestParseCtyValueNumber(t *testing.T) {
	val, err := ParseCtyValue(cty.NumberIntVal(42), []string{"number"})
	if err != nil {
		t.Fatal(err)
	}

	if val != int32(42) {
		t.Fatalf("expected int32(42), got %v (%T)", val, val)
	}
}

func TestParseCtyValueBool(t *testing.T) {
	val, err := ParseCtyValue(cty.BoolVal(true), []string{"bool"})
	if err != nil {
		t.Fatal(err)
	}

	if val != true {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestParseCtyValueTuple(t *testing.T) {
	tuple := cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")})
	val, err := ParseCtyValue(tuple, []string{"tuple"})
	if err != nil {
		t.Fatal(err)
	}
	list, ok := val.([]string)

	if !ok {
		t.Fatalf("expected []string, got %T", val)
	}

	if len(list) != 2 || list[0] != "a" || list[1] != "b" {
		t.Fatalf("expected [a b], got %v", list)
	}
}

func TestParseCtyValueDynamic(t *testing.T) {
	val, err := ParseCtyValue(cty.DynamicVal, []string{"dynamic"})
	if err != nil {
		t.Fatal(err)
	}

	if val != nil {
		t.Fatalf("expected nil for dynamic, got %v", val)
	}
}

func TestParseCtyValueTypeRestriction(t *testing.T) {
	_, err := ParseCtyValue(cty.StringVal("hello"), []string{"number", "bool"})

	if err == nil {
		t.Fatal("expected type restriction error")
	}
}

func TestParseCtyValueFloatFallback(t *testing.T) {
	// Use a float that doesn't fit cleanly in int32
	val, err := ParseCtyValue(cty.NumberFloatVal(3.14), []string{"number"})
	if err != nil {
		t.Fatal(err)
	}
	f, ok := val.(float32)

	if !ok {
		t.Fatalf("expected float32, got %T (%v)", val, val)
	}

	if f < 3.13 || f > 3.15 {
		t.Fatalf("expected ~3.14, got %v", f)
	}
}
