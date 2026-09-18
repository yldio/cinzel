// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// A run that converted nothing used to return nil, so the shell saw a success
// and the output directory stayed empty. Pointing at the wrong directory lands
// exactly here, and so does a directory holding only files this tool has no
// definition to take out of. GitLab already reports it.
func TestUnparseThatConvertsNothingIsReported(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"a directory of empty documents": {"a.yaml": "# nothing here\n", "b.yaml": "\n"},
		"a single empty document":        {"a.yaml": "\n"},
	} {
		t.Run(name, func(t *testing.T) {
			err, out := unparseDir(t, files)

			if !errors.Is(err, errNoDefinitions) {
				t.Fatalf("expected errNoDefinitions, got %v", err)
			}

			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Error("nothing should have been written")
			}
		})
	}
}

// The guard must not fire on a run that did convert something.
func TestUnparseThatConvertsSomethingIsNotReported(t *testing.T) {
	err, out := unparseDir(t, map[string]string{
		"real.yml":  realWorkflow,
		"stray.yml": "# nothing here\n",
	})
	if err != nil {
		t.Fatalf("a run that converted a workflow should succeed, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(out, "real.hcl")); statErr != nil {
		t.Errorf("the real workflow was not converted: %v", statErr)
	}
}
