// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// runInit runs "cinzel init" with answers fed on stdin, against a config
// directory of its own, and returns the path of the file it wrote.
func runInit(t *testing.T, home string, answers string) string {
	t.Helper()

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	in, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := in.WriteString(answers); err != nil {
		t.Fatal(err)
	}

	if _, err := in.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	original := os.Stdin
	os.Stdin = in

	t.Cleanup(func() {
		os.Stdin = original
		_ = in.Close()
	})

	app := New(new(bytes.Buffer), "v.test")

	if err := app.Execute([]string{"cinzel", "init"}, []provider.Provider{}); err != nil {
		t.Fatalf("init: %v", err)
	}

	dir, err := configPath()
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Join(dir, "config.yaml")
}

// The overwrite prompt used to read stdin through a scanner of its own. The
// scanner that followed started with an empty buffer, so on a re-run the
// answer after it came back blank.
func TestInitRereadsAnswersAfterTheOverwritePrompt(t *testing.T) {
	home := t.TempDir()
	first := runInit(t, home, "anthropic\n")

	if got, err := os.ReadFile(first); err != nil {
		t.Fatal(err)
	} else if !strings.Contains(string(got), "default: anthropic") {
		t.Fatalf("first run lost its answer:\n%s", got)
	}

	// Same HOME, so the file is already there and the overwrite prompt runs.
	second := runInit(t, home, "y\nopenai\n")

	got, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(got), "default: openai") {
		t.Errorf("want the provider kept after the overwrite prompt, got:\n%s", got)
	}
}

// The config is written to disk and swept up by a backup of the home
// directory, so init does not put a key in it. The environment is the
// documented place for one, and it wins over a key added here by hand.
func TestInitWritesNoAPIKey(t *testing.T) {
	configFile := runInit(t, t.TempDir(), "anthropic\n")

	got, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(got), "api_key") {
		t.Errorf("init wrote an api_key field:\n%s", got)
	}

	for _, want := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("want %q named in the written config, got:\n%s", want, got)
		}
	}
}

// The command says the file is 0600. os.WriteFile only applies a mode when
// it creates the file, so a config left group-readable stayed that way while
// the message claimed otherwise. A config written before init stopped asking
// for keys may still hold one, so the mode still matters.
func TestInitTightensPermissionsOnAnExistingFile(t *testing.T) {
	// Windows has no Unix permission bits: os.Chmod only toggles the read-only
	// flag there, and a file reads back as 0666 whatever it was set to.
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on windows")
	}

	home := t.TempDir()
	configFile := runInit(t, home, "anthropic\n")

	if err := os.Chmod(configFile, 0644); err != nil {
		t.Fatal(err)
	}

	runInit(t, home, "y\nanthropic\n")

	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if mode := info.Mode().Perm(); mode != 0600 {
		t.Errorf("want mode 0600, got %#o", mode)
	}
}
