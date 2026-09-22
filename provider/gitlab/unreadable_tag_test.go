// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"strings"
	"testing"
)

// TestUnreadableTagIsRejected covers silent corruption. "!reference" is how a
// pipeline pulls a keyword out of another job, and neither decoder here knows
// the tag: both drop it and hand back the plain sequence under it, so a
// reference arrived as an ordinary list and was written to HCL as shell
// commands. The run exited 0 and the pipeline that came back was a different
// one.
func TestUnreadableTagIsRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, yaml string }{
		{
			"reference as a whole keyword",
			".setup:\n  after_script: [echo bye]\nbuild:\n  script: [make]\n  after_script: !reference [.setup, after_script]\n",
		},
		{
			"reference inside a script list",
			".setup:\n  script: [echo hi]\nbuild:\n  script:\n    - !reference [.setup, script]\n    - make\n",
		},
		{
			"reference deep in a rule",
			".rules:\n  rules:\n    - if: '$CI'\nbuild:\n  script: [make]\n  rules: !reference [.rules, rules]\n",
		},
		{
			"any other unknown tag",
			"build:\n  script: [make]\n  image: !custom alpine\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := unparseGitLab(t, tc.yaml)
			if err == nil {
				t.Fatal("expected the tag to be rejected, got no error")
			}

			if !errors.Is(err, errUnreadableYAMLTag) {
				t.Fatalf("expected errUnreadableYAMLTag, got: %v", err)
			}
		})
	}
}

// TestStandardTagsStillPass keeps the refusal above from taking the tags that
// only name the type of what is already there, which survives the trip.
func TestStandardTagsStillPass(t *testing.T) {
	t.Parallel()

	yaml := "stages: [build]\nbuild:\n  stage: build\n  script: !!seq [make]\n  variables:\n    A: !!str plain\n    B: !!int 3\n"

	out, err := unparseGitLab(t, yaml)
	if err != nil {
		t.Fatalf("standard tags should convert, got: %v", err)
	}

	if !strings.Contains(out, "make") {
		t.Fatalf("expected the script to survive, got:\n%s", out)
	}
}
