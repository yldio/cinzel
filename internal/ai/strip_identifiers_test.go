// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStripKeepsIdentifiers pins what the skeleton still carries, so the
// wording the assist command shows before sending it stays true.
func TestStripKeepsIdentifiers(t *testing.T) {
	dir := t.TempDir()

	hcl := `workflow "release_acme" {
  filename = "release"

  job "deploy_to_acme_prod" {
    acme_internal_flag = "on"
  }
}
`

	if err := os.WriteFile(filepath.Join(dir, "acme-release.hcl"), []byte(hcl), 0644); err != nil {
		t.Fatal(err)
	}

	result, _ := StripHCLContext(dir)

	for _, kept := range []string{
		"acme-release.hcl",
		"workflow",
		`"release_acme"`,
		`"deploy_to_acme_prod"`,
		"acme_internal_flag",
	} {
		if !strings.Contains(result, kept) {
			t.Errorf("expected %q in the skeleton, got:\n%s", kept, result)
		}
	}

	if strings.Contains(result, `"release"`) {
		t.Errorf("expected the filename value to be stripped, got:\n%s", result)
	}
}
