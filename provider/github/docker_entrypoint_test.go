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

const dockerAction = `name: My Action
description: does things
runs:
  using: docker
  image: Dockerfile
  pre-entrypoint: setup.sh
  entrypoint: main.sh
  post-entrypoint: cleanup.sh
`

// A docker action may run a script before and after its entrypoint. Neither
// key was in the schema, so a valid action.yml was rejected outright with
// 'unknown field "post-entrypoint"'.
func TestDockerEntrypointHooksRoundtrip(t *testing.T) {
	tmp := t.TempDir()

	// An action is named by the directory holding it, so it is written into
	// one rather than straight into the temporary directory.
	dir := filepath.Join(tmp, "hooked")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	in := filepath.Join(dir, "action.yml")

	if err := os.WriteFile(in, []byte(dockerAction), 0o644); err != nil {
		t.Fatal(err)
	}

	hclDir := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir}); err != nil {
		t.Fatalf("unparse: %v", err)
	}

	hclBytes, err := os.ReadFile(filepath.Join(hclDir, "hooked.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"pre_entrypoint", "post_entrypoint"} {
		if !strings.Contains(string(hclBytes), want) {
			t.Fatalf("want %q in HCL, got:\n%s", want, hclBytes)
		}
	}

	backIn := filepath.Join(tmp, "back.hcl")

	if err := os.WriteFile(backIn, hclBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	yamlDir := filepath.Join(tmp, "yaml")

	if err := New().Parse(provider.ProviderOps{File: backIn, OutputDirectory: yamlDir}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	var got []byte

	err = filepath.Walk(yamlDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".yml") {
			got, err = os.ReadFile(path)

			return err
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"pre-entrypoint: setup.sh", "post-entrypoint: cleanup.sh", "entrypoint: main.sh"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("want %q after the roundtrip, got:\n%s", want, got)
		}
	}
}

// An action without the hooks must not grow empty ones.
func TestDockerActionWithoutHooksIsUnchanged(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "plain")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	in := filepath.Join(dir, "action.yml")

	plain := `name: My Action
description: does things
runs:
  using: docker
  image: Dockerfile
  entrypoint: main.sh
`

	if err := os.WriteFile(in, []byte(plain), 0o644); err != nil {
		t.Fatal(err)
	}

	hclDir := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir}); err != nil {
		t.Fatalf("unparse: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(hclDir, "plain.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(got), "entrypoint") && strings.Contains(string(got), "pre_entrypoint") {
		t.Errorf("want no pre_entrypoint, got:\n%s", got)
	}
}
