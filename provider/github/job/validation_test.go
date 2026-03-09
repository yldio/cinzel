// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

import (
	"strings"
	"testing"
)

func TestValidateModel(t *testing.T) {
	t.Run("valid job with runs-on and steps", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			HasRunsOn: true,
			StepCount: 1,
		}, "runs-on")
		if err != nil {
			t.Fatalf("expected valid, got %v", err)
		}
	})

	t.Run("valid reusable workflow job", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			Uses: "org/repo/.github/workflows/ci.yml@main",
		}, "runs-on")
		if err != nil {
			t.Fatalf("expected valid, got %v", err)
		}
	})

	t.Run("uses with runs-on", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			Uses:      "org/repo/.github/workflows/ci.yml@main",
			HasRunsOn: true,
		}, "runs-on")

		if err == nil {
			t.Fatal("expected error for uses with runs-on")
		}
	})

	t.Run("uses with steps", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			Uses:      "org/repo/.github/workflows/ci.yml@main",
			StepCount: 1,
		}, "runs-on")

		if err == nil {
			t.Fatal("expected error for uses with steps")
		}
	})

	t.Run("with without uses", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			HasRunsOn: true,
			StepCount: 1,
			HasWith:   true,
		}, "runs-on")

		if err == nil {
			t.Fatal("expected error for with without uses")
		}
	})

	t.Run("secrets without uses", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			HasRunsOn:  true,
			StepCount:  1,
			HasSecrets: true,
		}, "runs-on")

		if err == nil {
			t.Fatal("expected error for secrets without uses")
		}
	})

	t.Run("missing runs-on without uses", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			StepCount: 1,
		}, "runs-on")

		if err == nil {
			t.Fatal("expected error for missing runs-on")
		}
	})

	t.Run("no steps without uses", func(t *testing.T) {
		err := ValidateModel(ValidationModel{
			HasRunsOn: true,
			StepCount: 0,
		}, "runs-on")

		if err == nil {
			t.Fatal("expected error for zero steps")
		}
	})
}

func TestModelFromYAMLStepsTypeValidation(t *testing.T) {
	_, err := ModelFromYAML("build", map[string]any{
		"runs-on": "ubuntu-latest",
		"steps":   map[string]any{"run": "echo hi"},
	})

	if err == nil {
		t.Fatal("expected steps type error")
	}

	if !strings.Contains(err.Error(), "'steps' must be a list") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNeedsFromYAML(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		needs, err := NeedsFromYAML(nil)
		if err != nil {
			t.Fatal(err)
		}

		if needs != nil {
			t.Fatalf("expected nil, got %v", needs)
		}
	})

	t.Run("single string", func(t *testing.T) {
		needs, err := NeedsFromYAML("build")
		if err != nil {
			t.Fatal(err)
		}

		if len(needs) != 1 || needs[0] != "build" {
			t.Fatalf("expected [build], got %v", needs)
		}
	})

	t.Run("list of strings", func(t *testing.T) {
		needs, err := NeedsFromYAML([]any{"build", "test"})
		if err != nil {
			t.Fatal(err)
		}

		if len(needs) != 2 {
			t.Fatalf("expected 2, got %d", len(needs))
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := NeedsFromYAML("")

		if err == nil {
			t.Fatal("expected error for empty string")
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		_, err := NeedsFromYAML(123)

		if err == nil {
			t.Fatal("expected error for invalid type")
		}
	})
}

func TestModelFromParsed(t *testing.T) {
	t.Run("basic job", func(t *testing.T) {
		p := Parsed{
			ID: "build",
			Body: map[string]any{
				"runs-on": "ubuntu-latest",
				"steps":   []any{map[string]any{"run": "echo hi"}},
			},
		}
		m, err := ModelFromParsed(p)
		if err != nil {
			t.Fatal(err)
		}

		if !m.HasRunsOn {
			t.Fatal("expected HasRunsOn true")
		}

		if m.StepCount != 1 {
			t.Fatalf("expected 1 step, got %d", m.StepCount)
		}
	})

	t.Run("reusable workflow job", func(t *testing.T) {
		p := Parsed{
			ID: "call",
			Body: map[string]any{
				"uses": "org/repo/.github/workflows/ci.yml@main",
			},
		}
		m, err := ModelFromParsed(p)
		if err != nil {
			t.Fatal(err)
		}

		if m.Uses != "org/repo/.github/workflows/ci.yml@main" {
			t.Fatalf("expected uses value, got %q", m.Uses)
		}
	})
}

func TestNeedsFromYAMLEmptyEntry(t *testing.T) {
	_, err := NeedsFromYAML([]any{"build", ""})

	if err == nil {
		t.Fatal("expected empty needs entry error")
	}

	if !strings.Contains(err.Error(), "non-empty strings") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateNeedsReferences(t *testing.T) {
	t.Run("duplicate need", func(t *testing.T) {
		err := ValidateNeedsReferences([]string{"build", "build"}, map[string]ValidationModel{
			"build": {ID: "build"},
		})

		if err == nil {
			t.Fatal("expected duplicate need error")
		}

		if !strings.Contains(err.Error(), "duplicate needed job") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unknown need", func(t *testing.T) {
		err := ValidateNeedsReferences([]string{"missing"}, map[string]ValidationModel{
			"build": {ID: "build"},
		})

		if err == nil {
			t.Fatal("expected unknown need error")
		}

		if !strings.Contains(err.Error(), "cannot find needed job") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestValidateNeedsCycles(t *testing.T) {
	t.Run("self cycle", func(t *testing.T) {
		err := ValidateNeedsCycles(map[string]ValidationModel{
			"build": {ID: "build", Needs: []string{"build"}},
		})

		if err == nil {
			t.Fatal("expected cycle error")
		}

		if !strings.Contains(err.Error(), "dependency cycle") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("transitive cycle", func(t *testing.T) {
		err := ValidateNeedsCycles(map[string]ValidationModel{
			"a": {ID: "a", Needs: []string{"b"}},
			"b": {ID: "b", Needs: []string{"c"}},
			"c": {ID: "c", Needs: []string{"a"}},
		})

		if err == nil {
			t.Fatal("expected cycle error")
		}

		if !strings.Contains(err.Error(), "dependency cycle") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("no cycle", func(t *testing.T) {
		err := ValidateNeedsCycles(map[string]ValidationModel{
			"a": {ID: "a"},
			"b": {ID: "b", Needs: []string{"a"}},
			"c": {ID: "c", Needs: []string{"a", "b"}},
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
}
