// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// An "extends" name that sanitizes to nothing was replaced with a made-up
// label rather than refused, so "." went out as template.template and bound to
// whatever template happened to carry that name. "needs" already refuses the
// same input; this is the matching refusal for "extends".
func TestExtendsNameThatSanitizesToNothingIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		yml  string
	}{
		{
			name: "no template of that name",
			yml: `stages:
  - build
a:
  stage: build
  script:
    - make
  extends:
    - "."
`,
		},
		{
			name: "binds to an unrelated template",
			yml: `stages:
  - build
".template":
  script:
    - make base
a:
  stage: build
  script:
    - make
  extends:
    - "."
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yml)
			if err == nil {
				t.Fatal("want an error naming the extends entry")
			}

			if !strings.Contains(err.Error(), "extends") {
				t.Fatalf("error does not mention extends: %v", err)
			}
		})
	}
}
