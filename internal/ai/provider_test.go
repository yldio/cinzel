// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go/v3"
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
		// A response cut off at the token limit ends mid-fence. The opening
		// line used to be handed to the YAML parser along with the document.
		{
			name:  "unclosed fence runs to the end",
			input: "```yaml\nname: test\non:\n  push:",
			want:  "name: test\non:\n  push:",
		},
		{
			name:  "unclosed bare fence",
			input: "Here you go:\n\n```\nname: test",
			want:  "name: test",
		},
		{
			name:  "a closed fence then a truncated one",
			input: "```yaml\nname: one\n```\n\n```yaml\nname: two",
			want:  "name: one\n---\nname: two",
		},
		{
			name:  "text after a closed fence is not a fence",
			input: "```yaml\nname: test\n```\n\nHope this helps!",
			want:  "name: test",
		},
		// A workflow that writes a fence of its own, inside a "run: |" block,
		// carries backticks at column 4. Taking those for a fence returned the
		// one line after the last of them and the workflow was gone.
		{
			name:  "an indented fence inside a run block is not one",
			input: "name: test\njobs:\n  build:\n    steps:\n      - run: |\n          cat <<EOF\n          ```sh\n          echo hi\n          ```\n          EOF\n",
			want:  "name: test\njobs:\n  build:\n    steps:\n      - run: |\n          cat <<EOF\n          ```sh\n          echo hi\n          ```\n          EOF",
		},
		{
			name:  "a wrapped workflow keeps the fence it writes",
			input: "```yaml\nname: test\njobs:\n  build:\n    steps:\n      - run: |\n          echo '```'\n```\n",
			want:  "name: test\njobs:\n  build:\n    steps:\n      - run: |\n          echo '```'",
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

// Both SDK error types print themselves out of their Request and Response, so
// a bare struct panics when classifyError falls through to the message. A real
// one always carries both.
func apiRequestAndResponse(status int) (*http.Request, *http.Response) {
	req := httptest.NewRequest(http.MethodPost, "https://api.example.com/v1/messages", nil)

	return req, &http.Response{StatusCode: status, Request: req}
}

func openaiError(status int) *openai.Error {
	req, res := apiRequestAndResponse(status)

	return &openai.Error{StatusCode: status, Request: req, Response: res}
}

func anthropicError(status int) *anthropic.Error {
	req, res := apiRequestAndResponse(status)

	return &anthropic.Error{StatusCode: status, Request: req, Response: res}
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
		// The status comes off the typed error, so a message that happens to
		// carry other digits does not decide the classification.
		{
			name:     "a 429 carried in a typed error",
			err:      openaiError(http.StatusTooManyRequests),
			contains: "rate limited",
		},
		{
			name:     "a 401 carried in a typed error",
			err:      anthropicError(http.StatusUnauthorized),
			contains: "invalid API key",
		},
		{
			name:     "a 402 carried in a typed error",
			err:      openaiError(http.StatusPaymentRequired),
			contains: "quota exceeded",
		},
		{
			name:     "a wrapped typed error is still read",
			err:      fmt.Errorf("openai API: %w", openaiError(http.StatusTooManyRequests)),
			contains: "rate limited",
		},
		{
			name:     "a status with no advice falls through",
			err:      openaiError(http.StatusInternalServerError),
			contains: "LLM API error",
		},
		// An SDK error prints its response body, so a server fault whose body
		// mentions a rate limit read as one under the old text search.
		{
			name:     "text is not consulted when the status is known",
			err:      fmt.Errorf("%w: rate_limit_exceeded", openaiError(http.StatusInternalServerError)),
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
		// With no newline to cut at, the limit lands wherever it lands. A cut
		// through a multibyte rune left a partial encoding behind, which is
		// not a rune and reaches the provider as U+FFFD.
		{
			name:   "a cut one byte into a three-byte rune",
			input:  "\u65e5\u672c\u8a9e",
			maxLen: 4,
			want:   "\u65e5",
		},
		{
			name:   "a cut two bytes into a three-byte rune",
			input:  "\u65e5\u672c\u8a9e",
			maxLen: 5,
			want:   "\u65e5",
		},
		{
			name:   "a cut on a rune boundary keeps every rune",
			input:  "\u65e5\u672c\u8a9e",
			maxLen: 6,
			want:   "\u65e5\u672c",
		},
		{
			name:   "a cut three bytes into a four-byte rune",
			input:  "a\U0001f642b",
			maxLen: 4,
			want:   "a",
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
