// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

import "testing"

func TestAxesFromMap(t *testing.T) {
	axes := AxesFromMap(map[string]any{"b": 2, "a": 1})

	if len(axes) != 2 {
		t.Fatalf("expected 2 axes, got %d", len(axes))
	}

	if axes[0].Name != "a" {
		t.Fatalf("expected sorted order, first key should be a, got %s", axes[0].Name)
	}
}

func TestNormalizeStrategyMatrixNoVariable(t *testing.T) {
	matrix := map[string]any{"os": []any{"linux"}}
	norm, err := NormalizeStrategyMatrix(matrix)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := norm["os"]; !ok {
		t.Fatal("expected os key to remain")
	}
}

func TestNormalizeStrategyMatrixDuplicate(t *testing.T) {
	matrix := map[string]any{
		"os":       []any{"linux"},
		"variable": []any{map[string]any{"name": "os", "value": []any{"darwin"}}},
	}
	_, err := NormalizeStrategyMatrix(matrix)

	if err == nil {
		t.Fatal("expected duplicate key error")
	}
}

func TestNormalizeStrategyMatrixInvalidVariable(t *testing.T) {
	matrix := map[string]any{
		"variable": "not-valid",
	}
	_, err := NormalizeStrategyMatrix(matrix)

	if err == nil {
		t.Fatal("expected error for invalid variable type")
	}
}

func TestNormalizeStrategyMatrix(t *testing.T) {
	t.Run("name-value list", func(t *testing.T) {
		matrix := map[string]any{
			"variable": []any{
				map[string]any{"name": "os", "value": []any{"linux", "darwin"}},
				map[string]any{"name": "arch", "value": []any{"amd64"}},
			},
		}

		norm, err := NormalizeStrategyMatrix(matrix)
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := norm["variable"]; ok {
			t.Fatalf("expected variable key to be removed, got %#v", norm)
		}

		if _, ok := norm["os"]; !ok {
			t.Fatalf("expected os axis key, got %#v", norm)
		}
	})

	t.Run("map form", func(t *testing.T) {
		matrix := map[string]any{
			"variable": map[string]any{"os": []any{"linux"}, "arch": []any{"amd64"}},
		}

		norm, err := NormalizeStrategyMatrix(matrix)
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := norm["os"]; !ok {
			t.Fatalf("expected os axis key, got %#v", norm)
		}
	})
}
