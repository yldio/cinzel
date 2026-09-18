// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A key the HCL schema does not declare was written out as an attribute
// anyway, so unparse exited 0 on a file the same tool's parse direction then
// refused with "an argument named ... is not expected here". The refusal
// belongs where the input that caused it is still in hand.
func TestUnknownKeyIsRefused(t *testing.T) {
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
  bogus_keyword: x
`,
			want: "bogus_keyword",
		},
		{
			name: "default",
			yml: `stages: [build]
default:
  bogus_default: x
build:
  stage: build
  script: [make]
`,
			want: "bogus_default",
		},
		{
			name: "nested block",
			yml: `stages: [build]
build:
  stage: build
  script: [make]
  cache:
    paths: [vendor]
    bogus_cache: x
`,
			want: "bogus_cache",
		},
		{
			name: "rule",
			yml: `stages: [build]
build:
  stage: build
  script: [make]
  rules:
    - if: '$CI'
      bogus_rule: y
`,
			want: "bogus_rule",
		},
		{
			name: "workflow rule",
			yml: `stages: [build]
workflow:
  rules:
    - if: '$CI'
      bogus_rule: y
build:
  stage: build
  script: [make]
`,
			want: "bogus_rule",
		},
		{
			name: "need",
			yml: `stages: [build]
a:
  stage: build
  script: [make]
build:
  stage: build
  script: [make]
  needs:
    - job: a
      bogus_need: y
`,
			want: "bogus_need",
		},
		{
			name: "spec",
			yml: `spec:
  inputs:
    env: {}
  bogus_spec: x
---
stages: [build]
build:
  stage: build
  script: [make]
`,
			want: "bogus_spec",
		},
		{
			name: "variable",
			yml: `stages: [build]
variables:
  TOKEN:
    value: "x"
    bogus_var: y
build:
  stage: build
  script: [make]
`,
			want: "bogus_var",
		},
		{
			name: "top-level mapping read as a job",
			yml: `stages: [build]
build:
  stage: build
  script: [make]
unknown_top:
  a: 1
`,
			want: "a",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseYAMLString(t, tc.yml)
			if err == nil {
				t.Fatal("want an error naming the unknown key")
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error does not name the key: %v", err)
			}
		})
	}
}
