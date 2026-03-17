// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the maximum time to wait for an LLM response.
	DefaultTimeout = 120 * time.Second

	// DefaultMaxTokens is the maximum number of tokens in the LLM response.
	DefaultMaxTokens = 4096
)

var fencePattern = regexp.MustCompile("(?s)```(?:ya?ml)?\\s*\n(.*?)```")

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

func classifyError(err error, providerName string) error {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "authentication") || strings.Contains(msg, "401"):
		return fmt.Errorf("invalid API key for %s: %w", providerName, err)
	case strings.Contains(msg, "insufficient_quota") || strings.Contains(msg, "billing"):
		return fmt.Errorf("API quota exceeded for %s. Check your plan and billing at your provider's dashboard: %w", providerName, err)
	case strings.Contains(msg, "rate_limit") || strings.Contains(msg, "429"):
		return fmt.Errorf("API rate limited. Try again in a moment: %w", err)
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("LLM request timed out after %s. Try a simpler prompt", DefaultTimeout)
	default:
		return fmt.Errorf("LLM API error (%s): %w", providerName, err)
	}
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
func StripFences(s string) string {
	matches := fencePattern.FindAllStringSubmatch(s, -1)
	if len(matches) > 0 {
		var parts []string
		for _, m := range matches {
			parts = append(parts, strings.TrimSpace(m[1]))
		}

		return strings.Join(parts, "\n---\n")
	}

	return strings.TrimSpace(s)
}

// SystemPrompt returns the system prompt for the given CI provider name.
func SystemPrompt(providerName string) string {
	return fmt.Sprintf(`You are a CI/CD workflow generator for %s.

Your output will be converted to HCL where steps are reusable blocks shared across workflows and jobs. When generating multiple workflows, use IDENTICAL step definitions for common operations (checkout, setup, install dependencies, build, test). Give shared steps consistent names and IDs across all workflows so they can be deduplicated.

Generate valid %s YAML based on the user's description.

Rules:
- Output ONLY valid YAML. No markdown code fences, no explanations, no commentary.
- Use current action versions (tags like @v4, not SHAs).
- Set minimum required permissions.
- Use environment variables for secrets (e.g. secrets.MY_SECRET), never hardcode values.
- Include descriptive step names and IDs. Use consistent names: checkout, setup_go, install_deps, build, test, lint — not step_1, step_2.
- Follow %s best practices and conventions.
- When relevant, base your output on official starter workflows.
- If the request implies multiple workflows, separate them with --- (YAML document separator).
- Each YAML document should be a complete, valid workflow.
- When multiple workflows share steps (e.g. checkout + setup), use the EXACT same step name and id in each workflow so they can be deduplicated into a single reusable definition.`, providerName, providerName, providerName)
}
