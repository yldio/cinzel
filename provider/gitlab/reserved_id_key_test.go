// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// "id" is not a GitLab keyword: it is the attribute cinzel writes to record a
// job's name when the block label had to be sanitized. A job body carrying one
// was written straight into the block, where it overwrote that record, and the
// job came back under the wrong name with both directions exiting 0.
func TestJobKeyNamedIDIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
	}{
		{
			name: "job",
			yml: `stages: [build]
build:
  stage: build
  script: [make]
  id: shadow
`,
		},
		{
			name: "template",
			yml: `stages: [build]
.base:
  id: shadow
  script: [make]
build:
  stage: build
  extends: .base
  script: [make]
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yml)
			if err == nil {
				t.Fatal("want an error naming the reserved key")
			}

			if !strings.Contains(err.Error(), "'id'") {
				t.Fatalf("error does not name the key: %v", err)
			}
		})
	}
}
