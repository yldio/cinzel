// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// The guard on a passed-through top-level key asks whether sanitizing the key
// changes it. The empty key is the one key that is not an identifier and is
// unchanged by sanitizing, since there is nothing in it to change, so it
// passed and was written as an attribute with no name at all: a line reading
// `= "v"`, at exit 0, which cinzel's own parse then refused with "An argument
// or block definition is required here."
//
// See TestANonIdentifierTopLevelKeyIsRefused for the keys that guard already
// caught.
func TestAnEmptyTopLevelKeyIsRefused(t *testing.T) {
	const job = "build:\n  script:\n    - make\n"

	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			name: "a scalar under the empty key",
			yaml: job + `"": v` + "\n",
		},
		{
			name: "a list under the empty key",
			yaml: job + `"":` + "\n  - a\n  - b\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yaml)
			if err == nil {
				t.Fatal("want an error, got nil")
			}

			if !strings.Contains(err.Error(), "is not a valid HCL identifier") {
				t.Errorf("error does not say why the key was refused: %v", err)
			}
		})
	}
}
