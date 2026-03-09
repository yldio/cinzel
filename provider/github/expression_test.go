// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"strings"
	"testing"
)

func TestValidateExpressionSyntax(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{name: "plain string", input: "hello world"},
		{name: "valid expression", input: "${{ github.sha }}"},
		{name: "multiple expressions", input: "${{ github.workflow }} #${{ github.run_number }}"},
		{name: "expression in text", input: "echo ${{ github.ref }}"},
		{name: "unclosed expression", input: "${{ github.sha", wantErr: "unclosed expression"},
		{name: "empty expression", input: "${{  }}", wantErr: "empty expression"},
		{name: "nested braces ok", input: "${{ toJSON(github.event) }}"},
		{name: "plain braces no error", input: "obj = {a: 1}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExpressionSyntax("test", tt.input)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateExpressions(t *testing.T) {
	t.Run("valid workflow", func(t *testing.T) {
		wf := map[string]any{
			"name":     "CI",
			"run-name": "${{ github.workflow }} #${{ github.run_number }}",
			"jobs": map[string]any{
				"build": map[string]any{
					"runs-on": "${{ matrix.os }}",
					"steps": []any{
						map[string]any{"run": "echo ${{ github.sha }}"},
					},
				},
			},
		}

		if err := validateExpressions(wf); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unclosed in nested step", func(t *testing.T) {
		wf := map[string]any{
			"jobs": map[string]any{
				"build": map[string]any{
					"steps": []any{
						map[string]any{"run": "echo ${{ broken"},
					},
				},
			},
		}
		err := validateExpressions(wf)

		if err == nil {
			t.Fatal("expected error for unclosed expression")
		}

		if !strings.Contains(err.Error(), "unclosed expression") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
