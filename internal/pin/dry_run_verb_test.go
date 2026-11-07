// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A dry run reported each action in the past tense: "pinned
// actions/checkout@v4" over a file it had been told not to write. Under
// --check those lines are the whole report — the reader opens the failed step
// to see which actions it named — and every one of them said the fix was
// already in.
func TestDryRunDoesNotReportAChangeItDidNotMake(t *testing.T) {
	content := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

	sha := "abc123def456abc123def456abc123def456abc1"

	t.Run("pin", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "steps.hcl")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		resolver := &mockResolver{shas: map[string]string{"actions/checkout@v4": sha}}

		var buf bytes.Buffer

		if _, err := PinFile(context.Background(), path, resolver, &buf, true); err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(buf.String(), "would pin actions/checkout@v4") {
			t.Errorf("a dry run reported a write it did not make: %q", buf.String())
		}
	})

	t.Run("upgrade", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "steps.hcl")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		resolver := &stubUpgrader{
			latestTags: map[string]string{"actions/checkout": "v5"},
			shas:       map[string]string{"actions/checkout@v5": sha},
		}

		var buf bytes.Buffer

		if _, err := UpgradeFile(context.Background(), path, resolver, &buf, UpgradeOptions{DryRun: true}); err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(buf.String(), "would upgrade actions/checkout") {
			t.Errorf("a dry run reported a write it did not make: %q", buf.String())
		}
	})
}

// The control: a real run still reports in the past tense, because it did.
func TestARealRunStillReportsWhatItDid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "steps.hcl")

	content := `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{shas: map[string]string{
		"actions/checkout@v4": "abc123def456abc123def456abc123def456abc1",
	}}

	var buf bytes.Buffer

	if _, err := PinFile(context.Background(), path, resolver, &buf, false); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "pinned actions/checkout@v4") {
		t.Errorf("a real run should report what it did: %q", buf.String())
	}
}
