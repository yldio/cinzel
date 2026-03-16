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

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	defaultModel  = "claude-sonnet-4-5-20250514"
	apiTimeout    = 120 * time.Second
	apiKeyEnvVar  = "ANTHROPIC_API_KEY"
	maxTokens     = 4096
)

var errMissingAPIKey = errors.New(
	"ANTHROPIC_API_KEY environment variable is not set.\n\n" +
		"Set it with:\n" +
		"  export ANTHROPIC_API_KEY=sk-ant-...\n\n" +
		"Get your key at https://console.anthropic.com/settings/keys",
)

var errEmptyResponse = errors.New("LLM returned empty response. Try a more specific prompt")

// Generate calls the Anthropic API with the given system and user prompts,
// returning the text response.
func Generate(ctx context.Context, systemPrompt, userPrompt, model string) (string, error) {
	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return "", errMissingAPIKey
	}

	if model == "" {
		model = defaultModel
	}

	ctx, cancel := context.WithTimeout(ctx, apiTimeout)
	defer cancel()

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    model,
		MaxTokens: maxTokens,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return "", classifyError(err)
	}

	text := extractText(message)
	if strings.TrimSpace(text) == "" {
		return "", errEmptyResponse
	}

	return text, nil
}

func extractText(msg *anthropic.Message) string {
	var parts []string

	for _, block := range msg.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}

	return strings.Join(parts, "\n")
}

func classifyError(err error) error {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "authentication") || strings.Contains(msg, "401"):
		return fmt.Errorf("invalid API key. Set %s env var: %w", apiKeyEnvVar, err)
	case strings.Contains(msg, "rate_limit") || strings.Contains(msg, "429"):
		return fmt.Errorf("API rate limited. Try again in a moment: %w", err)
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("LLM request timed out after %s. Try a simpler prompt", apiTimeout)
	default:
		return fmt.Errorf("LLM API error: %w", err)
	}
}

// StripFences removes markdown code fences from LLM output, returning clean YAML.
func StripFences(s string) string {
	fencePattern := regexp.MustCompile("(?s)```(?:ya?ml)?\\s*\n(.*?)```")

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

Generate valid %s YAML based on the user's description.

Rules:
- Output ONLY valid YAML. No markdown code fences, no explanations, no commentary.
- Use current action versions (tags like @v4, not SHAs).
- Set minimum required permissions.
- Use environment variables for secrets (e.g. secrets.MY_SECRET), never hardcode values.
- Include descriptive step names and IDs.
- Follow %s best practices and conventions.
- When relevant, base your output on official starter workflows.
- If the request implies multiple workflows, separate them with --- (YAML document separator).
- Each YAML document should be a complete, valid workflow.`, providerName, providerName, providerName)
}
