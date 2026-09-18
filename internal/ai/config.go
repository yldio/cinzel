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

// configWarnings reports what is worth saying about a config file holding an
// api_key. cinzel init has not written that field since it stopped asking for
// keys, so a file with one in it predates that or was edited by hand.
//
// Two separate things are wrong with such a file, and they are said
// separately: the key is in a file at all, which is why the field is only a
// migration path, and the file may be one others can read, which is the
// urgent one. A file with no key in it is nobody else's business either way,
// so it says nothing.
func configWarnings(path string, cfg Config) []string {
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

	warnings := []string{fmt.Sprintf(
		"%s holds an api_key. Set the key in the environment instead and drop the field: it is read for configs written before cinzel init stopped asking for one, and that will not last.",
		path,
	)}

	// Windows has no Unix permission bits: a file reads back as 0666 whatever
	// it was set to, so the mode check would fire on every config there. The
	// key is still in a file, which is what the warning above is for.
	if runtime.GOOS == "windows" {
		return warnings
	}

	info, err := os.Stat(path)
	if err != nil {
		return warnings
	}

	mode := info.Mode().Perm()
	if mode&0077 == 0 {
		return warnings
	}

	return append(warnings, fmt.Sprintf(
		"%s is readable by others (mode %#o) while it holds that key. Run chmod 600 on it.",
		path, mode,
	))
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
