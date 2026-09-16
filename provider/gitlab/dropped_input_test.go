// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"io"
	"os"
	"strings"
	"testing"
)

// The HCL workflow block holds name, auto_cancel and rules. Anything else was
// dropped in silence, although the top-level loop warns about a key it cannot
// place.
func TestUnknownWorkflowKeyIsReported(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
		want string
	}{
		{
			name: "an unknown key",
			yml:  "workflow:\n  name: pipe\n  something_else: kept\n",
			want: "something_else",
		},
		{
			name: "the three known keys are quiet",
			yml:  "workflow:\n  name: pipe\n  auto_cancel:\n    on_new_commit: interruptible\n  rules:\n    - if: $CI\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			yml := tc.yml + "build:\n  script:\n    - echo hi\n"

			warnings := captureStderr(t, func() {
				if err := unparseYAMLString(t, yml); err != nil {
					t.Fatalf("Unparse() error = %v", err)
				}
			})

			if tc.want == "" {
				if warnings != "" {
					t.Fatalf("want no warning, got %q", warnings)
				}

				return
			}

			if !strings.Contains(warnings, tc.want) {
				t.Errorf("want %q named in %q", tc.want, warnings)
			}
		})
	}
}

// validateServices ran for each job and for "default", but not for the
// top-level services config.go also accepts.
func TestTopLevelServicesAreValidated(t *testing.T) {
	for _, tc := range []struct {
		name     string
		services string
		wantErr  string
	}{
		{
			name:     "an object with no name",
			services: `[{ alias = "db" }]`,
			wantErr:  "must include 'name'",
		},
		{
			name:     "an empty string",
			services: `[""]`,
			wantErr:  "non-empty strings",
		},
		{
			name:     "a sound service is unaffected",
			services: `[{ name = "postgres:16" }]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, "services = "+tc.services+"\n\njob \"build\" {\n  script = [\"echo hi\"]\n}\n")

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

// captureStderr returns whatever fn wrote to os.Stderr. The warnings under test
// go there rather than into an error, so there is nothing else to assert on.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	original := os.Stderr
	os.Stderr = w

	defer func() { os.Stderr = original }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	return string(out)
}
