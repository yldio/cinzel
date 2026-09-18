// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go/v3"
)

const (
	// DefaultTimeout is the maximum time to wait for an LLM response.
	DefaultTimeout = 120 * time.Second

	// DefaultMaxTokens is the maximum number of tokens in the LLM response.
	DefaultMaxTokens = 4096
)

// Both patterns anchor a fence line to column 0. Matching one anywhere took
// the backticks a workflow writes inside a "run: |" block as a fence of its
// own: a response with no wrapper at all, holding a fenced README, came back
// as the one line after the last of them and the workflow was gone.
var fencePattern = regexp.MustCompile("(?ms)^```(?:ya?ml)?[ \\t]*\n(.*?)^```[ \\t]*$")

// openFencePattern matches a fence that is never closed, which is what a
// response cut off at the token limit looks like.
var openFencePattern = regexp.MustCompile("(?ms)\\A.*?^```(?:ya?ml)?[ \\t]*\n(.*)\\z")

// GenerateRequest holds the parameters for an LLM generation call.
type GenerateRequest struct {
	SystemPrompt string
	UserPrompt   string
	Model        string
}

// GenerateResponse holds the LLM response text and token usage.
type GenerateResponse struct {
	Text         string
	InputTokens  int64
	OutputTokens int64
}

// TotalTokens returns the sum of input and output tokens.
func (r GenerateResponse) TotalTokens() int64 {
	return r.InputTokens + r.OutputTokens
}

// Provider defines the interface for LLM providers.
type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	Name() string
}

// GenerateWithTimeout calls the provider with a timeout and validates the response.
func GenerateWithTimeout(ctx context.Context, p Provider, req GenerateRequest) (GenerateResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	response, err := p.Generate(ctx, req)
	if err != nil {
		return GenerateResponse{}, classifyError(err, p.Name())
	}

	if strings.TrimSpace(response.Text) == "" {
		return GenerateResponse{}, errEmptyResponse
	}

	return response, nil
}

// classifyError turns a provider error into advice the user can act on.
//
// Both SDKs wrap an HTTP failure in a typed error carrying the status code, so
// that is read first. Searching the message text for "401" or "429" got this
// wrong in both directions: a model name or a prompt echoed back in the message
// could carry those digits, and a status that never reached the text was
// missed.
//
// The text search stays as a fallback, for a provider error that is not one of
// the two SDK types.
func classifyError(err error, providerName string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("LLM request timed out after %s. Try a simpler prompt", DefaultTimeout)
	}

	msg := err.Error()

	// A known status is the answer, and the text is not consulted at all: an
	// SDK error prints its response body, so a 500 whose body happens to
	// mention a rate limit would otherwise be reported as one.
	if status, ok := apiStatusCode(err); ok {
		switch status {
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("invalid API key for %s. Check your API key is correct", providerName)
		case http.StatusPaymentRequired:
			return fmt.Errorf("API quota exceeded for %s. Check your plan and billing at your provider's dashboard", providerName)
		case http.StatusTooManyRequests:
			return fmt.Errorf("API rate limited. Try again in a moment")
		default:
			return fmt.Errorf("LLM API error (%s): %s", providerName, msg)
		}
	}

	switch {
	case strings.Contains(msg, "authentication") || strings.Contains(msg, "401"):
		return fmt.Errorf("invalid API key for %s. Check your API key is correct", providerName)
	case strings.Contains(msg, "insufficient_quota") || strings.Contains(msg, "billing"):
		return fmt.Errorf("API quota exceeded for %s. Check your plan and billing at your provider's dashboard", providerName)
	case strings.Contains(msg, "rate_limit") || strings.Contains(msg, "429"):
		return fmt.Errorf("API rate limited. Try again in a moment")
	default:
		return fmt.Errorf("LLM API error (%s): %s", providerName, msg)
	}
}

// apiStatusCode reads the HTTP status out of whichever SDK raised the error.
// The two types are unrelated, so each is unwrapped in turn.
func apiStatusCode(err error) (int, bool) {
	var anthropicErr *anthropic.Error

	if errors.As(err, &anthropicErr) {
		return anthropicErr.StatusCode, true
	}

	var openaiErr *openai.Error

	if errors.As(err, &openaiErr) {
		return openaiErr.StatusCode, true
	}

	return 0, false
}

// resolveAPIKey returns the provided key, or falls back to the environment
// variable. Returns missingErr if neither is set.
func resolveAPIKey(provided, envVar string, missingErr error) (string, error) {
	if provided != "" {
		return provided, nil
	}

	if key := os.Getenv(envVar); key != "" {
		return key, nil
	}

	return "", missingErr
}

// StripFences removes markdown code fences from LLM output, returning clean YAML.
//
// An opening fence with no closer runs to the end of the text. A response cut
// off at the token limit ends that way, and it used to be returned verbatim, so
// the YAML parser was handed the "```yaml" line along with the document.
func StripFences(s string) string {
	matches := fencePattern.FindAllStringSubmatch(s, -1)
	if len(matches) > 0 {
		var parts []string
		for _, m := range matches {
			parts = append(parts, strings.TrimSpace(m[1]))
		}

		// A closed fence may still be followed by an unclosed one, when the
		// response was cut off partway through a second document.
		if tail := trailingOpenFence(s, matches); tail != "" {
			parts = append(parts, tail)
		}

		return strings.Join(parts, "\n---\n")
	}

	if m := openFencePattern.FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(m[1])
	}

	return strings.TrimSpace(s)
}

// trailingOpenFence returns the content of an unclosed fence after the last
// closed one, or "" when there is none.
func trailingOpenFence(s string, matches [][]string) string {
	last := matches[len(matches)-1][0]

	idx := strings.LastIndex(s, last)
	if idx < 0 {
		return ""
	}

	m := openFencePattern.FindStringSubmatch(s[idx+len(last):])
	if m == nil {
		return ""
	}

	return strings.TrimSpace(m[1])
}

// SystemPrompt returns the system prompt for the given CI provider name.
func SystemPrompt(providerName string) string {
	return fmt.Sprintf(`You are a CI/CD workflow generator for %s.

Your output will be converted to HCL where steps are reusable blocks shared across workflows and jobs. When generating multiple workflows, use IDENTICAL step definitions for common operations (checkout, setup, install dependencies, build, test). Give shared steps consistent names and IDs across all workflows so they can be deduplicated.

Generate valid %s YAML based on the user's description.

Rules:
- Output ONLY valid YAML. No markdown code fences, no explanations, no commentary.
- For action versions, use the LATEST major version tag (e.g. actions/checkout@v6, actions/setup-go@v5). Versions will be automatically pinned to SHAs after generation.
- Set minimum required permissions.
- Use environment variables for secrets (e.g. secrets.MY_SECRET), never hardcode values.
- Include descriptive step names and IDs. Use consistent names: checkout, setup_go, install_deps, build, test, lint — not step_1, step_2.
- Follow %s best practices and conventions.
- When relevant, base your output on official starter workflows.
- If the request implies multiple workflows, separate them with --- (YAML document separator).
- Each YAML document should be a complete, valid workflow.
- When multiple workflows share steps (e.g. checkout + setup), use the EXACT same step name and id in each workflow so they can be deduplicated into a single reusable definition.`, providerName, providerName, providerName)
}
