// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// Keys inside a generated block become bare HCL identifiers, and parse maps
// "_" back to "-". A key holding "_" used to come back holding "-" instead,
// and a key holding a dot used to produce HCL that does not parse.
func TestNestedBlockKeysThatCannotRoundtrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  string
		want string
	}{
		{"underscore", "my_custom_key", "read back as a dash"},
		{"dot", "a.b", "not a valid identifier"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			src := filepath.Join(tmp, "c.yaml")

			content := `on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
    container:
      image: node:18
      ` + tc.key + `: v
    steps:
      - run: x
`

			if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}

			err := New().Unparse(provider.ProviderOps{File: src, OutputDirectory: filepath.Join(tmp, "u")})

			if err == nil {
				out, _ := os.ReadFile(filepath.Join(tmp, "u", "c.hcl"))
				t.Fatalf("expected an error for key %q, got:\n%s", tc.key, out)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error does not name the reason, got: %v", err)
			}

			if !strings.Contains(err.Error(), tc.key) {
				t.Errorf("error does not name the key, got: %v", err)
			}
		})
	}
}

// Hyphenated keys are what the GitHub schema actually uses here, and they must
// keep working.
func TestNestedBlockHyphenatedKeysRoundtrip(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "c.yaml")

	content := `on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
    concurrency:
      group: g
      cancel-in-progress: true
    defaults:
      run:
        working-directory: ./src
    steps:
      - run: x
`

	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Unparse(provider.ProviderOps{File: src, OutputDirectory: filepath.Join(tmp, "u")}); err != nil {
		t.Fatal(err)
	}

	if err := p.Parse(provider.ProviderOps{File: filepath.Join(tmp, "u", "c.hcl"), OutputDirectory: filepath.Join(tmp, "p")}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(tmp, "p", "c.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"cancel-in-progress: true", "working-directory: ./src"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("lost %q:\n%s", want, got)
		}
	}
}
