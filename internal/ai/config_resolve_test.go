// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import "testing"

func TestResolveAPIKeyOrder(t *testing.T) {
	withConfig := Config{Providers: map[string]ProviderConfig{
		"anthropic": {APIKey: "anthropic-from-file"},
		"openai":    {APIKey: "openai-from-file"},
	}}

	tests := []struct {
		name     string
		cfg      Config
		provider string
		env      map[string]string
		want     string
	}{
		{
			name:     "env var beats the config file",
			cfg:      withConfig,
			provider: "anthropic",
			env:      map[string]string{anthropicAPIKeyEnvVar: "anthropic-from-env"},
			want:     "anthropic-from-env",
		},
		{
			name:     "config file is used when the env var is unset",
			cfg:      withConfig,
			provider: "anthropic",
			want:     "anthropic-from-file",
		},
		{
			name:     "an empty env var does not shadow the config file",
			cfg:      withConfig,
			provider: "anthropic",
			env:      map[string]string{anthropicAPIKeyEnvVar: ""},
			want:     "anthropic-from-file",
		},
		{
			name:     "each provider reads its own env var",
			cfg:      withConfig,
			provider: "openai",
			env: map[string]string{
				anthropicAPIKeyEnvVar: "anthropic-from-env",
				openaiAPIKeyEnvVar:    "openai-from-env",
			},
			want: "openai-from-env",
		},
		{
			name:     "the anthropic env var does not leak into openai",
			cfg:      Config{},
			provider: "openai",
			env:      map[string]string{anthropicAPIKeyEnvVar: "anthropic-from-env"},
			want:     "",
		},
		{
			name:     "an empty name resolves to the default provider",
			cfg:      Config{},
			provider: "",
			env:      map[string]string{anthropicAPIKeyEnvVar: "anthropic-from-env"},
			want:     "anthropic-from-env",
		},
		{
			name:     "an unknown provider reads no env var",
			cfg:      Config{},
			provider: "gemini",
			env:      map[string]string{anthropicAPIKeyEnvVar: "anthropic-from-env"},
			want:     "",
		},
		{
			name:     "an unknown provider still reads its config entry",
			cfg:      Config{Providers: map[string]ProviderConfig{"gemini": {APIKey: "gemini-from-file"}}},
			provider: "gemini",
			want:     "gemini-from-file",
		},
		{
			name:     "no key anywhere resolves to empty",
			cfg:      Config{},
			provider: "anthropic",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(anthropicAPIKeyEnvVar, "")
			t.Setenv(openaiAPIKeyEnvVar, "")

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			if got := tt.cfg.ResolveAPIKey(tt.provider); got != tt.want {
				t.Errorf("ResolveAPIKey(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}
