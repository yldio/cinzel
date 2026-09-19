// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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
		// A shell script closing nested JSON was read as an orphaned closer as
		// soon as the same string also held a real expression.
		{name: "json then an expression", input: `echo '{"a": {"b": 1}}' && echo ${{ github.sha }}`},
		{name: "an expression then json", input: `echo ${{ github.sha }} && echo '{"a": {"b": 1}}'`},
		{name: "json with no expression", input: `echo '{"a": {"b": 1}}'`},
		// GitHub reads a "}}" that no "${{" opened as plain text, and actionlint
		// 1.7.12 accepts every one of these. The check that called them
		// orphaned refused workflows GitHub runs, on the way in and on the way
		// back out.
		{name: "a closer with nothing to close", input: "echo }} && echo ${{ github.sha }}"},
		{name: "a closer after the expression closed", input: "${{ github.sha }} }}"},
		{name: "a lone closer in a plain string", input: "echo }}"},
		{name: "a shell brace beside an expression", input: `echo "${A}}" && echo ${{ github.sha }}`},
		{name: "format escaping its own braces", input: "${{ format('{{ {0} }}', github.sha) }}"},
		{name: "a closer doubled after an expression", input: "${{ github.sha }}}}"},
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

	// An attribute carrying an inline comment used to be wrapped, and the
	// wrapper made walkStrings skip its value entirely.
	t.Run("commented attribute is still validated", func(t *testing.T) {
		wf := map[string]any{
			"env": map[string]any{
				"URL": annotated{value: "${{ broken", comment: "# staging only"},
			},
		}

		if err := validateExpressions(plainMap(wf)); err == nil {
			t.Fatal("expected error for unclosed expression behind a comment")
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
