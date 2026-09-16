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

// Detection used to require a "name" as well as a "runs", so an action without
// one fell through to the step-only path and came back with "not a valid type"
// rather than the field it was missing.
func TestActionIsDetectedWithoutAName(t *testing.T) {
	for _, tc := range []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "an action with no name",
			yaml:    "description: does a thing\nruns:\n  using: node20\n  main: index.js\n",
			wantErr: "action must define 'name'",
		},
		{
			name:    "an action with no using",
			yaml:    "name: Thing\nruns:\n  main: index.js\n",
			wantErr: "runs.using",
		},
		{
			name: "a complete action is unaffected",
			yaml: "name: Thing\ndescription: does a thing\nruns:\n  using: node20\n  main: index.js\n",
		},
		{
			// No "runs", so this is still read as a map of steps.
			name: "a step-only document is still detected",
			yaml: "checkout:\n  uses: actions/checkout@v4\ntest:\n  run: go test ./...\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "in.yaml")

			if err := os.WriteFile(path, []byte(tc.yaml), 0o600); err != nil {
				t.Fatal(err)
			}

			err := New().Unparse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("want an error, got nil")
			}

			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want %q in %q", tc.wantErr, err)
			}
		})
	}
}
