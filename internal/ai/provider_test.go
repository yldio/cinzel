// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStripFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no fences",
			input: "name: test\non:\n  push:",
			want:  "name: test\non:\n  push:",
		},
		{
			name:  "yaml fence",
			input: "```yaml\nname: test\non:\n  push:\n```",
			want:  "name: test\non:\n  push:",
		},
		{
			name:  "yml fence",
			input: "```yml\nname: test\n```",
			want:  "name: test",
		},
		{
			name:  "bare fence",
			input: "```\nname: test\n```",
			want:  "name: test",
		},
		{
			name:  "fence with surrounding text",
			input: "Here is the workflow:\n\n```yaml\nname: test\n```\n\nHope this helps!",
			want:  "name: test",
		},
		{
			name:  "multiple fences joined with separator",
			input: "```yaml\nname: workflow1\n```\n\n```yaml\nname: workflow2\n```",
			want:  "name: workflow1\n---\nname: workflow2",
		},
		{
			name:  "whitespace only input",
			input: "   \n\n  ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripFences(tt.input)
			if got != tt.want {
				t.Errorf("StripFences():\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "authentication error",
			err:      errors.New("authentication failed: 401 Unauthorized"),
			contains: "invalid API key",
		},
		{
			name:     "quota exceeded",
			err:      errors.New("insufficient_quota: check billing"),
			contains: "quota exceeded",
		},
		{
			name:     "rate limit",
			err:      errors.New("rate_limit_exceeded: 429"),
			contains: "rate limited",
		},
		{
			name:     "timeout",
			err:      context.DeadlineExceeded,
			contains: "timed out",
		},
		{
			name:     "generic error",
			err:      errors.New("something unexpected"),
			contains: "LLM API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyError(tt.err, "test-provider")
			if !strings.Contains(got.Error(), tt.contains) {
				t.Errorf("classifyError():\ngot:  %q\nwant to contain: %q", got.Error(), tt.contains)
			}
		})
	}
}

func TestResolveAPIKey(t *testing.T) {
	sentinel := errors.New("key missing")

	t.Run("provided key used", func(t *testing.T) {
		key, err := resolveAPIKey("my-key", "NONEXISTENT_VAR", sentinel)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if key != "my-key" {
			t.Errorf("expected my-key, got %s", key)
		}
	})

	t.Run("missing key returns sentinel", func(t *testing.T) {
		_, err := resolveAPIKey("", "NONEXISTENT_VAR_12345", sentinel)
		if !errors.Is(err, sentinel) {
			t.Errorf("expected sentinel error, got %v", err)
		}
	})
}

func TestTruncateAtNewline(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "no truncation needed",
			input:  "short",
			maxLen: 100,
			want:   "short",
		},
		{
			name:   "truncates at newline",
			input:  "line1\nline2\nline3",
			maxLen: 10,
			want:   "line1",
		},
		{
			name:   "no newline in range",
			input:  "abcdefghij",
			maxLen: 5,
			want:   "abcde",
		},
		{
			name:   "exact length",
			input:  "abc",
			maxLen: 3,
			want:   "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateAtNewline(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateAtNewline():\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
