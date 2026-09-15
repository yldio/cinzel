// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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

	// GitHub keeps adding scopes, and a list of the ones known when this was
	// written rejected workflows GitHub itself accepts.
	t.Run("scope added after this validator was written", func(t *testing.T) {
		for _, scope := range []string{"models", "vulnerability-alerts", "artifact-metadata", "code-quality"} {
			if err := ValidatePermissions(map[string]any{scope: "read"}); err != nil {
				t.Errorf("scope %q: %v", scope, err)
			}
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
