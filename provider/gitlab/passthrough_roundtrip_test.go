// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"
)

// A top-level key outside the schema is written out as a bare attribute, with
// a warning and exit 0. Nothing read it back, so the next parse refused the
// file unparse had just written: "An argument named 'my_extra' is not expected
// here", at exit 1. The README puts the two commands side by side, so a
// pipeline carrying one such key failed on the second step of the documented
// way in.
func TestAPassedThroughTopLevelKeyIsReadBack(t *testing.T) {
	const pipeline = "stages:\n  - build\nbuild:\n  stage: build\n  script:\n    - make\n"

	for _, tc := range []struct {
		name string
		yml  string
		want string
	}{
		{
			name: "a string",
			yml:  pipeline + "my_extra: kept\n",
			want: "my_extra: kept",
		},
		{
			name: "a number",
			yml:  pipeline + "my_extra: 5\n",
			want: "my_extra: 5",
		},
		{
			name: "a boolean",
			yml:  pipeline + "my_extra: true\n",
			want: "my_extra: true",
		},
		{
			name: "a list",
			yml:  pipeline + "my_extra:\n  - a\n  - b\n",
			want: "my_extra:\n  - a\n  - b",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl, back := roundtripYAML(t, tc.yml)

			if !strings.Contains(back, tc.want) {
				t.Errorf("want %q back, got:\n%s\nfrom HCL:\n%s", tc.want, back, hcl)
			}
		})
	}
}

// The keys the schema does name are read by the blocks that declare them, so
// the passthrough reader must not find them a second time and overwrite what
// those blocks produced.
func TestTheSchemaKeysAreNotReadTwice(t *testing.T) {
	_, back := roundtripYAML(t, "stages:\n  - build\nimage: alpine\nbuild:\n  stage: build\n  script:\n    - make\n")

	if got := strings.Count(back, "image:"); got != 1 {
		t.Errorf("want one image key, got %d:\n%s", got, back)
	}

	if !strings.Contains(back, "stages:\n  - build") {
		t.Errorf("the stage list did not survive:\n%s", back)
	}
}
