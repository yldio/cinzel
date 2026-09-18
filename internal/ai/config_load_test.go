// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// configDirAt points os.UserConfigDir() at a directory of the test's own and
// returns the config file path LoadConfig will look for.
func configDirAt(t *testing.T) string {
	t.Helper()

	home := t.TempDir()

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	dir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "cinzel"), 0o750); err != nil {
		t.Fatal(err)
	}

	return filepath.Join(dir, "cinzel", "config.yaml")
}

// A config that exists and was not used used to be silent, which looked exactly
// like having no config at all: a file the process cannot read, or one with a
// typo in it, came out as every setting falling back to its default with
// nothing said.
func TestLoadConfigWarnsWhenTheFileIsUnusable(t *testing.T) {
	for _, tc := range []struct {
		name  string
		body  string
		mode  os.FileMode
		want  string
		skip  bool
		cfgOK bool
	}{
		{
			name: "no config at all says nothing",
			want: "",
		},
		{
			name: "a file that will not parse",
			body: "ai:\n  providers:\n   - this is not a mapping\n",
			mode: 0o600,
			want: "not valid YAML",
		},
		{
			// Running as root reads it anyway, and Windows has no mode bits.
			name: "a file that cannot be read",
			body: withoutKey,
			mode: 0o000,
			want: "could not be read",
			skip: runtime.GOOS == "windows" || os.Geteuid() == 0,
		},
		{
			name:  "a config that reads fine",
			body:  withoutKey,
			mode:  0o600,
			want:  "",
			cfgOK: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip("mode bits are not meaningful here")
			}

			path := configDirAt(t)

			if tc.body != "" {
				if err := os.WriteFile(path, []byte(tc.body), tc.mode); err != nil {
					t.Fatal(err)
				}

				if err := os.Chmod(path, tc.mode); err != nil {
					t.Fatal(err)
				}
			}

			cfg, warnings := LoadConfig()

			if tc.want == "" {
				if len(warnings) != 0 {
					t.Fatalf("LoadConfig() warned %q, want no warning", warnings)
				}
			} else {
				if len(warnings) != 1 {
					t.Fatalf("LoadConfig() warned %q, want one warning about %q", warnings, tc.want)
				}

				if !strings.Contains(warnings[0], tc.want) {
					t.Errorf("LoadConfig() warned %q, want it to mention %q", warnings[0], tc.want)
				}

				if !strings.Contains(warnings[0], path) {
					t.Errorf("LoadConfig() warned %q, want it to name %q", warnings[0], path)
				}
			}

			if tc.cfgOK && cfg.Providers["anthropic"].Model == "" {
				t.Errorf("LoadConfig() returned %+v, want the config it read", cfg)
			}
		})
	}
}
