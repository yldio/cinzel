// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"strings"
	"testing"

	"github.com/yldio/cinzel/internal/hclparser"
)

// The name becomes a YAML key. An empty one goes out as `"": v`, which
// actionlint reads as `string should not be empty` and GitHub refuses, and the
// job and workflow level path already said so.
func TestAnEmptyNameIsRefused(t *testing.T) {
	nameExpr := expression(t, `""`)
	valueExpr := expression(t, `"v"`)

	envCfg := EnvListConfig{{Name: nameExpr, Value: valueExpr}}

	if _, err := envCfg.Parse(hclparser.NewHCLVars()); err == nil {
		t.Error("expected an env block with an empty name to be refused")
	} else if !strings.Contains(err.Error(), "name must not be empty") {
		t.Errorf("env error does not say the name was empty: %v", err)
	}

	withCfg := WithListConfig{{Name: nameExpr, Value: valueExpr}}

	if _, err := withCfg.Parse(hclparser.NewHCLVars()); err == nil {
		t.Error("expected a with block with an empty name to be refused")
	} else if !strings.Contains(err.Error(), "name must not be empty") {
		t.Errorf("with error does not say the name was empty: %v", err)
	}
}

// Two blocks writing one key kept the last and dropped the first without a
// word, so a step quietly ran with a value its author never wrote. The job and
// workflow level path refuses the same thing with "two 'env' blocks both write
// 'A'".
func TestTwoNamesThatAreTheSameAreRefused(t *testing.T) {
	first := expression(t, `"A"`)
	second := expression(t, `"A"`)
	one := expression(t, `"1"`)
	two := expression(t, `"2"`)

	envCfg := EnvListConfig{{Name: first, Value: one}, {Name: second, Value: two}}

	if _, err := envCfg.Parse(hclparser.NewHCLVars()); err == nil {
		t.Error("expected two env blocks writing 'A' to be refused")
	} else if !strings.Contains(err.Error(), "A") {
		t.Errorf("env error does not name the key: %v", err)
	}

	withCfg := WithListConfig{{Name: first, Value: one}, {Name: second, Value: two}}

	if _, err := withCfg.Parse(hclparser.NewHCLVars()); err == nil {
		t.Error("expected two with blocks writing 'A' to be refused")
	} else if !strings.Contains(err.Error(), "A") {
		t.Errorf("with error does not name the key: %v", err)
	}
}

// Two different names are what a step normally carries, and both have to
// survive.
func TestTwoNamesThatDifferAreKept(t *testing.T) {
	cfg := EnvListConfig{
		{Name: expression(t, `"A"`), Value: expression(t, `"1"`)},
		{Name: expression(t, `"B"`), Value: expression(t, `"2"`)},
	}

	val, err := cfg.Parse(hclparser.NewHCLVars())
	if err != nil {
		t.Fatalf("expected both env blocks to be kept, got %v", err)
	}

	mapping := val.AsValueMap()

	if len(mapping) != 2 {
		t.Fatalf("expected two keys, got %d", len(mapping))
	}

	if mapping["A"].AsString() != "1" || mapping["B"].AsString() != "2" {
		t.Errorf("lost a value: %#v", mapping)
	}
}
