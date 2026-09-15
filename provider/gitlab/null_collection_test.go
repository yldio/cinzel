// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// GitLab lets "rules", "artifacts", "only" and "except" be written as null to
// clear what a job would otherwise inherit. Unparse used to reject the first
// two outright and lose the others, so valid YAML either failed or changed
// meaning.
func TestNullCollectionsSurvive(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "job rules",
			yaml: "job1:\n  script:\n    - make\n  rules:\n",
			want: "rules: null",
		},
		{
			name: "job artifacts",
			yaml: "job1:\n  script:\n    - make\n  artifacts:\n",
			want: "artifacts: null",
		},
		{
			name: "job cache",
			yaml: "job1:\n  script:\n    - make\n  cache:\n",
			want: "cache: null",
		},
		{
			name: "job only",
			yaml: "job1:\n  script:\n    - make\n  only:\n",
			want: "only: null",
		},
		{
			name: "job except",
			yaml: "job1:\n  script:\n    - make\n  except:\n",
			want: "except: null",
		},
		{
			name: "workflow rules",
			yaml: "workflow:\n  rules:\njob1:\n  script:\n    - make\n",
			want: "rules: null",
		},
		{
			name: "template rules",
			yaml: ".tpl:\n  rules:\n  script:\n    - make\njob1:\n  extends: .tpl\n  script:\n    - make\n",
			want: "rules: null",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripYAML(t, tc.yaml)

			if !strings.Contains(back, tc.want) {
				t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", tc.want, hcl, back)
			}
		})
	}
}

// An empty "extends" clears an inherited one, so it is not the same as
// leaving the keyword out.
func TestEmptyExtendsSurvives(t *testing.T) {
	hcl, back := roundtripYAML(t, "job1:\n  script:\n    - make\n  extends: []\n")

	if !strings.Contains(back, "extends: []") {
		t.Errorf("roundtrip lost the empty extends\nHCL:\n%s\nYAML:\n%s", hcl, back)
	}
}

// A null must not be invented where the keyword was absent, and must not
// replace one carrying a real value. Only the keywords GitLab allows a null
// on keep it; everywhere else a null stays dropped, since writing one back
// would emit YAML GitLab's own schema rejects.
func TestNullIsNotInvented(t *testing.T) {
	for _, tc := range []struct {
		name    string
		yaml    string
		notWant []string
	}{
		{
			name:    "absent keywords stay absent",
			yaml:    "job1:\n  script:\n    - make\n",
			notWant: []string{"rules:", "artifacts:", "only:", "except:", "cache:", "extends:"},
		},
		{
			name:    "populated rules keep their entries",
			yaml:    "job1:\n  script:\n    - make\n  rules:\n    - if: $CI_COMMIT_BRANCH\n",
			notWant: []string{"rules: null"},
		},
		{
			name:    "populated artifacts keep their entries",
			yaml:    "job1:\n  script:\n    - make\n  artifacts:\n    paths:\n      - out\n",
			notWant: []string{"artifacts: null"},
		},
		{
			name:    "populated extends keeps its entries",
			yaml:    ".tpl:\n  script:\n    - make\njob1:\n  extends: .tpl\n  script:\n    - make\n",
			notWant: []string{"extends: []"},
		},
		{
			name:    "a null image is not written back",
			yaml:    "job1:\n  script:\n    - make\n  image:\n",
			notWant: []string{"image:"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripYAML(t, tc.yaml)

			for _, notWant := range tc.notWant {
				if strings.Contains(back, notWant) {
					t.Errorf("roundtrip invented %q\nHCL:\n%s\nYAML:\n%s", notWant, hcl, back)
				}
			}
		})
	}
}
