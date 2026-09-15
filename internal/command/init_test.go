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
// scanner that followed started with an empty buffer, so on a re-run every
// answer came back blank and the keys were written as "".
func TestInitRereadsAnswersAfterTheOverwritePrompt(t *testing.T) {
	home := t.TempDir()
	first := runInit(t, home, "anthropic\nsk-ant-one\nsk-oa-one\n")

	if got, err := os.ReadFile(first); err != nil {
		t.Fatal(err)
	} else if !strings.Contains(string(got), "sk-ant-one") {
		t.Fatalf("first run lost its answers:\n%s", got)
	}

	// Same HOME, so the file is already there and the overwrite prompt runs.
	second := runInit(t, home, "y\nopenai\nsk-ant-two\nsk-oa-two\n")

	got, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"default: openai", "sk-ant-two", "sk-oa-two"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("want %q kept after the overwrite prompt, got:\n%s", want, got)
		}
	}

	if strings.Contains(string(got), `api_key: ""`) {
		t.Errorf("an answered key was written empty:\n%s", got)
	}
}

// The file holds API keys and the command says it is 0600. os.WriteFile only
// applies a mode when it creates the file, so a config left group-readable
// stayed that way while the message claimed otherwise.
func TestInitTightensPermissionsOnAnExistingFile(t *testing.T) {
	// Windows has no Unix permission bits: os.Chmod only toggles the read-only
	// flag there, and a file reads back as 0666 whatever it was set to.
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on windows")
	}

	home := t.TempDir()
	configFile := runInit(t, home, "anthropic\nsk-ant-one\nsk-oa-one\n")

	if err := os.Chmod(configFile, 0644); err != nil {
		t.Fatal(err)
	}

	runInit(t, home, "y\nanthropic\nsk-ant-two\nsk-oa-two\n")

	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if mode := info.Mode().Perm(); mode != 0600 {
		t.Errorf("want mode 0600, got %#o", mode)
	}
}
