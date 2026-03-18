// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	openaiDefaultModel = "gpt-4o"
	openaiAPIKeyEnvVar = "OPENAI_API_KEY"
)

// OpenAI implements the Provider interface using the OpenAI API.
type OpenAI struct {
	apiKey string
}

// NewOpenAI creates an OpenAI provider, reading the API key from the
// environment or the provided key string.
func NewOpenAI(apiKey string) (*OpenAI, error) {
	key, err := resolveAPIKey(apiKey, openaiAPIKeyEnvVar, errMissingOpenAIKey)
	if err != nil {
		return nil, err
	}

	return &OpenAI{apiKey: key}, nil
}

// Name returns the provider name.
func (o *OpenAI) Name() string {
	return "openai"
}

// Generate calls the OpenAI Chat Completions API.
func (o *OpenAI) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = openaiDefaultModel
	}

	client := openai.NewClient(option.WithAPIKey(o.apiKey))

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(req.SystemPrompt),
			openai.UserMessage(req.UserPrompt),
		},
	})
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("openai API: %w", err)
	}

	var text string
	if len(completion.Choices) > 0 {
		text = completion.Choices[0].Message.Content
	}

	return GenerateResponse{
		Text:         text,
		InputTokens:  completion.Usage.PromptTokens,
		OutputTokens: completion.Usage.CompletionTokens,
	}, nil
}
