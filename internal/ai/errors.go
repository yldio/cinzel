// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"errors"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

var (
	errEmptyResponse       = errors.New("LLM returned empty response. Try a more specific prompt")
	errMissingAnthropicKey = cinzelerror.UserInput(errors.New(
		"ANTHROPIC_API_KEY environment variable is not set.\n\n" +
			"Set it with:\n" +
			"  export ANTHROPIC_API_KEY=sk-ant-...\n\n" +
			"Get your key at https://console.anthropic.com/settings/keys",
	))
	errMissingOpenAIKey = cinzelerror.UserInput(errors.New(
		"OPENAI_API_KEY environment variable is not set.\n\n" +
			"Set it with:\n" +
			"  export OPENAI_API_KEY=sk-...\n\n" +
			"Get your key at https://platform.openai.com/api-keys",
	))
)
