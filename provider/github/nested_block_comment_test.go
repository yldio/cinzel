// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A key holding a nested map becomes a block, and the recursion that opens it
// writes only what is inside. The head comment was therefore written for a key
// holding a scalar and dropped for one holding a map, one line apart in the
// same file.
func TestANestedBlockKeepsTheCommentAboveIt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "above a nested map key",
			yaml: "on: push\n" +
				"jobs:\n" +
				"  build:\n" +
				"    runs-on: ubuntu-latest\n" +
				"    defaults:\n" +
				"      # why bash\n" +
				"      run:\n" +
				"        shell: bash\n" +
				"    steps:\n" +
				"      - run: echo hi\n",
			want: "# why bash\n    run {",
		},
		{
			// The control: a scalar under the same block already worked, and
			// still has to.
			name: "above a scalar key in the same block",
			yaml: "on: push\n" +
				"jobs:\n" +
				"  build:\n" +
				"    runs-on: ubuntu-latest\n" +
				"    defaults:\n" +
				"      run:\n" +
				"        # why bash\n" +
				"        shell: bash\n" +
				"    steps:\n" +
				"      - run: echo hi\n",
			want: "# why bash\n      shell = \"bash\"",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hcl := unparse(t, tc.yaml)

			if !strings.Contains(hcl, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, hcl)
			}
		})
	}
}
