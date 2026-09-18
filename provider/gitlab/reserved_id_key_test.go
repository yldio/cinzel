// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// "id" on a job and "name" on a variable are not GitLab keywords: they are the
// attributes cinzel writes to record what the block was called when the label
// had to be sanitized. A body carrying one of them either overwrote that record,
// so the job came back under the wrong name, or was dropped where the writer
// filled the attribute in itself. Both directions exited 0 either way.
func TestJobKeyNamedIDIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
		want string
	}{
		{
			name: "job",
			yml: `stages: [build]
build:
  stage: build
  script: [make]
  id: shadow
`,
			want: "'id'",
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
			want: "'id'",
		},
		{
			name: "variable",
			yml: `stages: [build]
variables:
  TOKEN:
    value: x
    name: OTHER
build:
  stage: build
  script: [make]
`,
			want: "'name'",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yml)
			if err == nil {
				t.Fatal("want an error naming the reserved key")
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error does not name the key: %v", err)
			}
		})
	}
}
