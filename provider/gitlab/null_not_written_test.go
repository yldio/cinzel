// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A null is only a legal spelling on the keywords GitLab reads as clearing an
// inherited value, and parse drops one written anywhere else. Unparse wrote it
// all the same, so "image:" came out as "image = null" for cinzel's own parse
// to delete: the HCL from the first pass and the HCL from the second differed
// by a line that carried nothing.
func TestADroppedNullIsNotWritten(t *testing.T) {
	for _, tc := range []struct {
		name    string
		yaml    string
		notWant string
	}{
		{
			name:    "a job keyword",
			yaml:    "job1:\n  script:\n    - make\n  image:\n",
			notWant: "image",
		},
		{
			name:    "a default keyword",
			yaml:    "default:\n  image:\njob1:\n  script:\n    - make\n",
			notWant: "image",
		},
		{
			name:    "a workflow keyword",
			yaml:    "workflow:\n  name:\njob1:\n  script:\n    - make\n",
			notWant: "name",
		},
		{
			name:    "a rule keyword",
			yaml:    "job1:\n  script:\n    - make\n  rules:\n    - if: $CI\n      when:\n",
			notWant: "when",
		},
		{
			name:    "a need keyword",
			yaml:    "job1:\n  script:\n    - make\n  needs:\n    - job: job2\n      optional:\njob2:\n  script:\n    - make\n",
			notWant: "optional",
		},
		{
			name:    "a top-level keyword",
			yaml:    "image:\njob1:\n  script:\n    - make\n",
			notWant: "image",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, _ := roundtripYAML(t, tc.yaml)

			// hclwrite pads the attribute names of a block to a common width,
			// so the gap before "=" depends on the longest name beside it.
			flat := strings.Join(strings.Fields(hcl), " ")

			if strings.Contains(flat, tc.notWant+" = null") {
				t.Errorf("unparse wrote a null parse drops: %q\nHCL:\n%s", tc.notWant, hcl)
			}
		})
	}
}

// The keywords GitLab does read as clearing an inherited value keep their
// null, and the HCL holding it is what a second pass writes: the first pass
// used to emit lines parse deleted, so the two differed.
func TestHCLIsStableAcrossTwoPasses(t *testing.T) {
	const yml = `default:
  image:
  services:
workflow:
  name:
  rules:
build:
  script:
    - echo
  image:
  tags:
  rules:
  artifacts:
  cache:
  services:
  only:
  except:
`

	first, back := roundtripYAML(t, yml)
	second, _ := roundtripYAML(t, back)

	if first != second {
		t.Errorf("HCL changed on the second pass\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	// hclwrite pads the attribute names of a block to a common width, so the
	// gap before "=" depends on the longest name beside it.
	flat := strings.Join(strings.Fields(first), " ")

	for _, want := range []string{"rules", "artifacts", "cache", "services", "only", "except"} {
		if !strings.Contains(flat, want+" = null") {
			t.Errorf("a null GitLab allows was lost: %q\nHCL:\n%s", want, first)
		}
	}
}
