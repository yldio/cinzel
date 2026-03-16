// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	anthropicDefaultModel = "claude-sonnet-4-5-20250514"
	anthropicAPIKeyEnvVar = "ANTHROPIC_API_KEY"
)

var errMissingAnthropicKey = errors.New(
	"ANTHROPIC_API_KEY environment variable is not set.\n\n" +
		"Set it with:\n" +
		"  export ANTHROPIC_API_KEY=sk-ant-...\n\n" +
		"Get your key at https://console.anthropic.com/settings/keys",
)

// Anthropic implements the Provider interface using the Anthropic API.
type Anthropic struct {
	apiKey string
}

// NewAnthropic creates an Anthropic provider, reading the API key from the
// environment or the provided key string.
func NewAnthropic(apiKey string) (*Anthropic, error) {
	if apiKey == "" {
		apiKey = os.Getenv(anthropicAPIKeyEnvVar)
	}

	if apiKey == "" {
		return nil, errMissingAnthropicKey
	}

	return &Anthropic{apiKey: apiKey}, nil
}

// Name returns the provider name.
func (a *Anthropic) Name() string {
	return "anthropic"
}

// Generate calls the Anthropic Messages API.
func (a *Anthropic) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	model := req.Model
	if model == "" {
		model = anthropicDefaultModel
	}

	client := anthropic.NewClient(option.WithAPIKey(a.apiKey))

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: DefaultMaxTokens,
		System: []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.UserPrompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic API: %w", err)
	}

	return extractAnthropicText(message), nil
}

func extractAnthropicText(msg *anthropic.Message) string {
	var parts []string

	for _, block := range msg.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}

	return strings.Join(parts, "\n")
}
