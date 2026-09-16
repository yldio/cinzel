// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A passed-through top-level key is written as an attribute name, so one that
// is not an identifier produced HCL nothing can read back — "my weird key = v"
// is three block labels and an equals sign. It went out with only the
// passthrough warning and exit 0.
func TestANonIdentifierTopLevelKeyIsRefused(t *testing.T) {
	const job = "build:\n  script:\n    - make\n"

	for _, tc := range []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "spaces in the key",
			yaml:    job + "my weird key: v\n",
			wantErr: `"my weird key" is not a valid HCL identifier`,
		},
		{
			name:    "a dash in the key",
			yaml:    job + "a-b: 1\n",
			wantErr: `"a-b" is not a valid HCL identifier`,
		},
		{
			name:    "a colon in the key",
			yaml:    job + "\"x:y\": v\n",
			wantErr: `"x:y" is not a valid HCL identifier`,
		},
		{
			name: "an identifier passes through as before",
			yaml: job + "some_extra: v\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yaml)

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
