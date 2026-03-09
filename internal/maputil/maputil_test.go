// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package maputil

import "testing"

func TestToStringAnyMap(t *testing.T) {
	t.Run("string map", func(t *testing.T) {
		in := map[string]any{"a": 1}
		out, ok := ToStringAnyMap(in)

		if !ok || out["a"] != 1 {
			t.Fatalf("expected successful string map conversion, got ok=%v out=%#v", ok, out)
		}
	})

	t.Run("any map", func(t *testing.T) {
		in := map[any]any{"a": 1}
		out, ok := ToStringAnyMap(in)

		if !ok || out["a"] != 1 {
			t.Fatalf("expected successful any map conversion, got ok=%v out=%#v", ok, out)
		}
	})

	t.Run("invalid key type", func(t *testing.T) {
		in := map[any]any{1: "a"}
		_, ok := ToStringAnyMap(in)

		if ok {
			t.Fatal("expected conversion failure for non-string key")
		}
	})
}

func TestSortedKeys(t *testing.T) {
	keys := SortedKeys(map[string]int{"b": 2, "a": 1, "c": 3})

	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("unexpected sorted keys: %#v", keys)
	}
}
