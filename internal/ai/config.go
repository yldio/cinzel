// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the AI section of the cinzel config file.
type Config struct {
	Default   string                    `yaml:"default"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig holds per-provider settings from the config file.
type ProviderConfig struct {
	Model  string `yaml:"model"`
	APIKey string `yaml:"api_key"`
}

type configFile struct {
	AI Config `yaml:"ai"`
}

// LoadConfig reads the cinzel config file from os.UserConfigDir()/cinzel/config.yaml.
// Returns an empty Config (not an error) if the file doesn't exist.
func LoadConfig() Config {
	dir, err := os.UserConfigDir()
	if err != nil {
		return Config{}
	}

	path := filepath.Join(dir, "cinzel", "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}
	}

	var cfg configFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}
	}

	return cfg.AI
}

// ResolveProviderName returns the provider name to use, applying the
// resolution order: CLI flag > config default > "anthropic".
func (c Config) ResolveProviderName(cliFlag string) string {
	if cliFlag != "" {
		return cliFlag
	}

	if c.Default != "" {
		return c.Default
	}

	return "anthropic"
}

// ResolveAPIKey returns the API key for the given provider, applying the
// resolution order: env var > config file.
func (c Config) ResolveAPIKey(providerName string) string {
	pc, ok := c.Providers[providerName]
	if ok && pc.APIKey != "" {
		return pc.APIKey
	}

	return ""
}

// ResolveModel returns the model for the given provider, applying the
// resolution order: CLI flag > config file > provider default.
func (c Config) ResolveModel(providerName, cliFlag string) string {
	if cliFlag != "" {
		return cliFlag
	}

	pc, ok := c.Providers[providerName]
	if ok && pc.Model != "" {
		return pc.Model
	}

	return ""
}
