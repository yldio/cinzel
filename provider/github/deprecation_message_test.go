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

// GitHub spells this input key in camel case, alone among the input keys.
// Parse wrote "deprecation-message", which GitHub ignores and cinzel's own
// strict shape rejects, so the generated file did not come back.
func TestDeprecationMessageRoundtrips(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "act.hcl")

	hcl := `action "myact" {
  filename = "myact"
  name     = "My Action"

  input "old_thing" {
    description         = "a thing"
    deprecation_message = "use new_thing"
  }

  runs {
    using = "node20"
    main  = "index.js"
  }
}
`

	if err := os.WriteFile(in, []byte(hcl), 0o600); err != nil {
		t.Fatal(err)
	}

	yamlDir := filepath.Join(tmp, "yaml")

	if err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: yamlDir}); err != nil {
		t.Fatalf("parse: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(yamlDir, "myact", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(got), "deprecationMessage: use new_thing") {
		t.Fatalf("want the camel-case key, got:\n%s", got)
	}

	// Unparse is also the strict shape check: it rejects a key the action
	// schema does not declare, which is how the wrong spelling showed up.
	hclDir := filepath.Join(tmp, "hcl")

	if err := New().Unparse(provider.ProviderOps{Directory: yamlDir, Recursive: true, OutputDirectory: hclDir}); err != nil {
		t.Fatalf("unparse: %v", err)
	}

	back, err := os.ReadFile(filepath.Join(hclDir, "myact.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	// The HCL attribute is snake case, which is what the schema declares. A
	// literal "deprecationMessage" attribute would not parse back.
	if !strings.Contains(string(back), `deprecation_message = "use new_thing"`) {
		t.Fatalf("want the snake-case attribute, got:\n%s", back)
	}

	againDir := filepath.Join(tmp, "yaml2")

	if err := New().Parse(provider.ProviderOps{File: filepath.Join(hclDir, "myact.hcl"), OutputDirectory: againDir}); err != nil {
		t.Fatalf("reparse: %v", err)
	}

	again, err := os.ReadFile(filepath.Join(againDir, "myact", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if string(again) != string(got) {
		t.Errorf("roundtrip is not stable:\nfirst:\n%s\nsecond:\n%s", got, again)
	}
}
