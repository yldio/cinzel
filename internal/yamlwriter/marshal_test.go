// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package yamlwriter

import (
	"strings"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestMarshalPrimitiveAndNilPointer(t *testing.T) {
	type sample struct {
		Name   string  `yaml:"name"`
		Count  int     `yaml:"count"`
		Nested *string `yaml:"nested,omitempty"`
	}

	out, err := Marshal(sample{Name: "cinzel", Count: 2})
	if err != nil {
		t.Fatalf("marshal should succeed: %v", err)
	}

	content := string(out)

	if !strings.Contains(content, "name: cinzel") {
		t.Fatalf("expected name in yaml output, got: %q", content)
	}

	if !strings.Contains(content, "count: 2") {
		t.Fatalf("expected count in yaml output, got: %q", content)
	}

	if strings.Contains(content, "nested") {
		t.Fatalf("expected nil pointer field to be omitted, got: %q", content)
	}
}

func TestConvertCtyValue(t *testing.T) {
	type withCty struct {
		Val cty.Value `yaml:"val"`
	}

	result, err := Convert(withCty{Val: cty.StringVal("hello")})
	if err != nil {
		t.Fatal(err)
	}

	m, ok := result.(map[string]any)

	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if m["val"] != "hello" {
		t.Fatalf("expected hello, got %v", m["val"])
	}
}

func TestConvertCtyNilVal(t *testing.T) {
	type withCty struct {
		Val cty.Value `yaml:"val"`
	}

	result, err := Convert(withCty{Val: cty.NilVal})
	if err != nil {
		t.Fatal(err)
	}

	m, ok := result.(map[string]any)

	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if _, exists := m["val"]; exists {
		t.Fatal("expected nil cty.Value to be omitted")
	}
}

func TestConvertSlice(t *testing.T) {
	result, err := Convert([]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}

	list, ok := result.([]any)

	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(list))
	}
}

func TestConvertMap(t *testing.T) {
	result, err := Convert(map[string]int{"a": 1})
	if err != nil {
		t.Fatal(err)
	}

	m, ok := result.(map[any]any)

	if !ok {
		t.Fatalf("expected map[any]any, got %T", result)
	}

	if m["a"] != 1 {
		t.Fatalf("expected 1, got %v", m["a"])
	}
}

func TestConvertDashTag(t *testing.T) {
	type sample struct {
		Visible string `yaml:"visible"`
		Hidden  string `yaml:"-"`
	}

	result, err := Convert(sample{Visible: "yes", Hidden: "no"})
	if err != nil {
		t.Fatal(err)
	}

	m, ok := result.(map[string]any)

	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if _, exists := m["Hidden"]; exists {
		t.Fatal("expected hidden field to be omitted")
	}

	if _, exists := m["-"]; exists {
		t.Fatal("expected dash-tagged field to be omitted")
	}
}

func TestConvertPointer(t *testing.T) {
	type inner struct {
		Name string `yaml:"name"`
	}

	v := &inner{Name: "ptr"}
	result, err := Convert(v)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := result.(map[string]any)

	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if m["name"] != "ptr" {
		t.Fatalf("expected ptr, got %v", m["name"])
	}
}
