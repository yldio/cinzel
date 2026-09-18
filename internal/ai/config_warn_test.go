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
		want int
	}{
		// The key is in a file, and the file is open: two things, said twice.
		{"a key others can read warns twice", withKey, 0644, 2},
		{"group-readable counts too", withKey, 0640, 2},
		// Tight permissions answer the second warning, not the first: the
		// field is a migration path, so a key in it is still worth saying.
		{"a key only the owner can read still warns", withKey, 0600, 1},
		{"no key, no business of ours", withoutKey, 0644, 0},
		{"no key and tight is fine", withoutKey, 0600, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.body, tt.mode)

			cfg := configOf(t, tt.body)
			got := configWarnings(path, cfg)

			if len(got) != tt.want {
				t.Fatalf("want %d warnings for mode %#o, got %d: %q", tt.want, tt.mode, len(got), got)
			}
		})
	}
}

// A warning names the file and what to do. It must never name the key, which
// is the thing a warning is most likely to be pasted into a bug report with.
func TestConfigWarningKeepsTheKeyOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on windows")
	}

	path := writeConfig(t, withKey, 0644)

	got := configWarnings(path, configOf(t, withKey))
	if len(got) != 2 {
		t.Fatalf("want two warnings, got %q", got)
	}

	for _, warning := range got {
		if strings.Contains(warning, "sk-ant-secret") {
			t.Errorf("the warning leaks the key: %q", warning)
		}

		if !strings.Contains(warning, path) {
			t.Errorf("want the path named, got %q", warning)
		}
	}
}

// A file cinzel cannot stat has no mode to judge, so the permission warning is
// not guessed at. The key is in the config either way, so that one still runs.
func TestConfigWarningsOnAMissingFile(t *testing.T) {
	got := configWarnings(filepath.Join(t.TempDir(), "gone.yaml"), configOf(t, withKey))
	if len(got) != 1 {
		t.Fatalf("want only the api_key warning for a missing file, got %q", got)
	}

	if strings.Contains(got[0], "readable by others") {
		t.Errorf("a file with no mode to read should not be called open: %q", got[0])
	}
}
