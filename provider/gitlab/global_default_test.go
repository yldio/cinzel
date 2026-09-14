// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// GitLab still reads "image", "before_script", "after_script", "cache" and
// "services" at the top level, where each means what the same key means under
// "default". Unparse used to write them back as bare attributes the HCL schema
// had no field for, so cinzel could not read its own output.
func TestGlobalDefaultKeywordsSurvive(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want []string
	}{
		{
			name: "image",
			yaml: "image: alpine\njob1:\n  script:\n    - make\n",
			want: []string{"image: alpine"},
		},
		{
			name: "before_script",
			yaml: "before_script:\n  - setup\njob1:\n  script:\n    - make\n",
			want: []string{"before_script:", "- setup"},
		},
		{
			name: "after_script",
			yaml: "after_script:\n  - teardown\njob1:\n  script:\n    - make\n",
			want: []string{"after_script:", "- teardown"},
		},
		{
			name: "cache",
			yaml: "cache:\n  key: shared\n  paths:\n    - vendor\njob1:\n  script:\n    - make\n",
			want: []string{"cache:", "key: shared", "- vendor"},
		},
		{
			name: "services",
			yaml: "services:\n  - postgres\njob1:\n  script:\n    - make\n",
			want: []string{"services:", "- postgres"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripYAML(t, tc.yaml)

			for _, want := range tc.want {
				if !strings.Contains(back, want) {
					t.Errorf("roundtrip lost %q\nHCL:\n%s\nYAML:\n%s", want, hcl, back)
				}
			}

			if !strings.Contains(back, "job1:") {
				t.Errorf("roundtrip lost the job\nHCL:\n%s\nYAML:\n%s", hcl, back)
			}
		})
	}
}

// A global keyword stays where it was written. GitLab does not document which
// wins when a pipeline has both, so folding one into the other would be a
// guess at the pipeline's meaning.
func TestGlobalDefaultKeywordsAreNotFolded(t *testing.T) {
	hcl, back := roundtripYAML(t,
		"image: alpine\ndefault:\n  image: debian\njob1:\n  script:\n    - make\n")

	if !strings.Contains(back, "image: alpine") || !strings.Contains(back, "image: debian") {
		t.Errorf("roundtrip lost one of the two images\nHCL:\n%s\nYAML:\n%s", hcl, back)
	}
}

// A global "cache" is a mapping, and must not be mistaken for a job.
func TestGlobalCacheIsNotAJob(t *testing.T) {
	hcl, back := roundtripYAML(t,
		"cache:\n  key: shared\n  paths:\n    - vendor\njob1:\n  script:\n    - make\n")

	if strings.Contains(hcl, `job "cache"`) {
		t.Errorf("a global cache became a job\nHCL:\n%s\nYAML:\n%s", hcl, back)
	}
}
