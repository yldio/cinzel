// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A file that holds no pipeline is passed over, and that used to leave no trace
// at all: pointing unparse at a directory holding one pipeline and four other
// YAML files reported the same success as a run that converted every one of
// them. The GitHub provider has warned about this since
// "a-skipped-file-counted-as-no-failure"; GitLab did not.
func TestASkippedFileIsNamed(t *testing.T) {
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in")

	if err := os.MkdirAll(in, 0o755); err != nil {
		t.Fatal(err)
	}

	for name, body := range map[string]string{
		"a-not-a-pipeline.yml": "name: Dependabot-ish\nversion: 2\n",
		"z-good.yml":           goodPipeline,
	} {
		if err := os.WriteFile(filepath.Join(in, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out := filepath.Join(tmp, "out")

	var err error

	warnings := captureStderr(t, func() {
		err = New().Unparse(provider.ProviderOps{Directory: in, OutputDirectory: out})
	})

	// Skipping is not failing: the run goes on and writes the pipeline it did
	// recognise.
	if err != nil {
		t.Fatalf("want the run to succeed, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(out, "z-good.hcl")); statErr != nil {
		t.Errorf("want the pipeline written: %v", statErr)
	}

	if !strings.Contains(warnings, "a-not-a-pipeline.yml") {
		t.Errorf("want the skipped file named on stderr, got: %q", warnings)
	}

	// The file that converted is not a skip, and saying so would read as one.
	if strings.Contains(warnings, "z-good.yml") {
		t.Errorf("want only the skipped file named, got: %q", warnings)
	}
}
