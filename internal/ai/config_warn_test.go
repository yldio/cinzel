// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// writeConfig writes a config file with the given mode and returns its path.
func writeConfig(t *testing.T, body string, mode os.FileMode) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}

	return path
}

// configOf decodes a config body the way LoadConfig does.
func configOf(t *testing.T, body string) Config {
	t.Helper()

	var cfg configFile

	if err := yaml.Unmarshal([]byte(body), &cfg); err != nil {
		t.Fatal(err)
	}

	return cfg.AI
}

const withKey = "ai:\n  providers:\n    anthropic:\n      api_key: sk-ant-secret\n"
const withoutKey = "ai:\n  providers:\n    anthropic:\n      model: claude-sonnet-4-5-20250514\n"

func TestConfigWarnings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on windows")
	}

	tests := []struct {
		name string
		body string
		mode os.FileMode
		want bool
	}{
		{"a key others can read warns", withKey, 0644, true},
		{"group-readable counts too", withKey, 0640, true},
		{"a key only the owner can read is fine", withKey, 0600, false},
		{"no key, no business of ours", withoutKey, 0644, false},
		{"no key and tight is fine", withoutKey, 0600, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.body, tt.mode)

			cfg := configOf(t, tt.body)
			got := configWarnings(path, cfg)

			if tt.want && len(got) == 0 {
				t.Fatalf("want a warning for mode %#o, got none", tt.mode)
			}

			if !tt.want && len(got) != 0 {
				t.Fatalf("want no warning for mode %#o, got %q", tt.mode, got)
			}
		})
	}
}

// The warning names the file and what to do. It must never name the key.
func TestConfigWarningKeepsTheKeyOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on windows")
	}

	path := writeConfig(t, withKey, 0644)

	got := configWarnings(path, configOf(t, withKey))
	if len(got) != 1 {
		t.Fatalf("want one warning, got %q", got)
	}

	if strings.Contains(got[0], "sk-ant-secret") {
		t.Errorf("the warning leaks the key: %q", got[0])
	}

	if !strings.Contains(got[0], path) {
		t.Errorf("want the path named, got %q", got[0])
	}
}

// A file cinzel cannot stat says nothing rather than guessing.
func TestConfigWarningsOnAMissingFile(t *testing.T) {
	got := configWarnings(filepath.Join(t.TempDir(), "gone.yaml"), configOf(t, withKey))
	if len(got) != 0 {
		t.Fatalf("want no warning for a missing file, got %q", got)
	}
}
