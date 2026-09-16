// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

// DefaultModels returns the model each provider uses when the config names
// none, keyed by provider name. cinzel init writes the template from this, so
// a new config starts on whatever the code already defaults to.
func DefaultModels() map[string]string {
	return map[string]string{
		"anthropic": anthropicDefaultModel,
		"openai":    openaiDefaultModel,
	}
}

// LoadConfig reads the cinzel config file from os.UserConfigDir()/cinzel/config.yaml.
// Returns an empty Config (not an error) if the file doesn't exist, and a
// warning for each thing worth saying about the file it did read.
func LoadConfig() (Config, []string) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, nil
	}

	path := filepath.Join(dir, "cinzel", "config.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, nil
	}

	var cfg configFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, nil
	}

	return cfg.AI, configWarnings(path, cfg.AI)
}

// configWarnings reports a config file that others can read while it holds a
// key. cinzel init writes 0600 and no key, so this only fires on a file that
// predates that or was edited by hand. A file with no key in it is nobody
// else's business either way, so it says nothing.
func configWarnings(path string, cfg Config) []string {
	// Windows has no Unix permission bits: a file reads back as 0666 whatever
	// it was set to, so the check would warn about every config there.
	if runtime.GOOS == "windows" {
		return nil
	}

	hasKey := false

	for _, pc := range cfg.Providers {
		if pc.APIKey != "" {
			hasKey = true

			break
		}
	}

	if !hasKey {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	mode := info.Mode().Perm()
	if mode&0077 == 0 {
		return nil
	}

	return []string{fmt.Sprintf(
		"%s holds an api_key and is readable by others (mode %#o). Run chmod 600 on it, or drop the key and set the API key in the environment instead.",
		path, mode,
	)}
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
	if envVar := apiKeyEnvVar(providerName); envVar != "" {
		if key := os.Getenv(envVar); key != "" {
			return key
		}
	}

	pc, ok := c.Providers[providerName]
	if ok && pc.APIKey != "" {
		return pc.APIKey
	}

	return ""
}

// apiKeyEnvVar returns the environment variable holding the API key for the
// given provider, or an empty string for a provider that has none. The empty
// name resolves to the default provider, matching ResolveProviderName.
func apiKeyEnvVar(providerName string) string {
	switch strings.ToLower(providerName) {
	case "anthropic", "":
		return anthropicAPIKeyEnvVar
	case "openai":
		return openaiAPIKeyEnvVar
	default:
		return ""
	}
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
