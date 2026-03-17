// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	anthropicDefaultModel = "claude-sonnet-4-5-20250514"
	anthropicAPIKeyEnvVar = "ANTHROPIC_API_KEY"
)

// Anthropic implements the Provider interface using the Anthropic API.
type Anthropic struct {
	apiKey string
}

// NewAnthropic creates an Anthropic provider, reading the API key from the
// environment or the provided key string.
func NewAnthropic(apiKey string) (*Anthropic, error) {
	key, err := resolveAPIKey(apiKey, anthropicAPIKeyEnvVar, errMissingAnthropicKey)
	if err != nil {
		return nil, err
	}

	return &Anthropic{apiKey: key}, nil
}

// Name returns the provider name.
func (a *Anthropic) Name() string {
	return "anthropic"
}

// Generate calls the Anthropic Messages API.
func (a *Anthropic) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
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
		return GenerateResponse{}, fmt.Errorf("anthropic API: %w", err)
	}

	return GenerateResponse{
		Text:         extractAnthropicText(message),
		InputTokens:  message.Usage.InputTokens,
		OutputTokens: message.Usage.OutputTokens,
	}, nil
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
