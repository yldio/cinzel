// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	openaiDefaultModel = "gpt-4o"
	openaiAPIKeyEnvVar = "OPENAI_API_KEY"
)

var errMissingOpenAIKey = errors.New(
	"OPENAI_API_KEY environment variable is not set.\n\n" +
		"Set it with:\n" +
		"  export OPENAI_API_KEY=sk-...\n\n" +
		"Get your key at https://platform.openai.com/api-keys",
)

// OpenAI implements the Provider interface using the OpenAI API.
type OpenAI struct {
	apiKey string
}

// NewOpenAI creates an OpenAI provider, reading the API key from the
// environment or the provided key string.
func NewOpenAI(apiKey string) (*OpenAI, error) {
	if apiKey == "" {
		apiKey = os.Getenv(openaiAPIKeyEnvVar)
	}

	if apiKey == "" {
		return nil, errMissingOpenAIKey
	}

	return &OpenAI{apiKey: apiKey}, nil
}

// Name returns the provider name.
func (o *OpenAI) Name() string {
	return "openai"
}

// Generate calls the OpenAI Chat Completions API.
func (o *OpenAI) Generate(ctx context.Context, req GenerateRequest) (string, error) {
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
		return "", fmt.Errorf("openai API: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", nil
	}

	return completion.Choices[0].Message.Content, nil
}
