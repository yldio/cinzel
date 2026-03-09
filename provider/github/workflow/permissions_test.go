// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package workflow

import (
	"strings"
	"testing"
)

func TestValidatePermissions(t *testing.T) {
	t.Run("nil is valid", func(t *testing.T) {
		if err := ValidatePermissions(nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("read-all shorthand", func(t *testing.T) {
		if err := ValidatePermissions("read-all"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("write-all shorthand", func(t *testing.T) {
		if err := ValidatePermissions("write-all"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid shorthand", func(t *testing.T) {
		err := ValidatePermissions("admin")

		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "invalid permissions shorthand") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid scope map", func(t *testing.T) {
		err := ValidatePermissions(map[string]any{
			"contents":    "read",
			"deployments": "write",
			"issues":      "none",
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unknown scope", func(t *testing.T) {
		err := ValidatePermissions(map[string]any{
			"admin": "read",
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "unknown permissions scope") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		err := ValidatePermissions(map[string]any{
			"contents": "admin",
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "invalid permission level") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty map is valid", func(t *testing.T) {
		if err := ValidatePermissions(map[string]any{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		err := ValidatePermissions(123)

		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "must be a string or an object") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
