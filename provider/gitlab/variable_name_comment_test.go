// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A variable carrying nothing but a value collapses to one scalar key, so every
// comment written inside the block belongs to that key. Only the value's line
// was read, so a note beside "name =" was dropped: the parse exited 0 and wrote
// the same bytes as before, which left nothing for "diff" to report either.
func TestAVariableNameCommentReachesTheKeyItNames(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want string
	}{
		{
			name: "beside the name",
			hcl: "variable \"V\" {\n" +
				"  name  = \"V\" # what this sets\n" +
				"  value = \"x\"\n}\n",
			want: "V: x # what this sets",
		},
		{
			name: "above the name",
			hcl: "variable \"V\" {\n" +
				"  # what this sets\n" +
				"  name  = \"V\"\n" +
				"  value = \"x\"\n}\n",
			want: "# what this sets\n  V: x",
		},
		{
			// The one line the key has belongs to the value, so the name's
			// stacks above rather than taking a line already spoken for.
			name: "beside both the name and the value",
			hcl: "variable \"V\" {\n" +
				"  name  = \"V\" # about the name\n" +
				"  value = \"x\" # about the value\n}\n",
			want: "# about the name\n  V: x # about the value",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := parseVariablePipeline(t, tc.hcl)

			if !strings.Contains(got, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, got)
			}
		})
	}
}

// parseVariablePipeline parses a pipeline holding the variable block and
// returns the YAML it wrote.
func parseVariablePipeline(t *testing.T, variable string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "in.hcl")
	output := filepath.Join(dir, "out")

	hcl := "stages = [\"build\"]\n\n" + variable + "\n" +
		"job \"build\" {\n  stage  = \"build\"\n  script = [\"make\"]\n}\n"

	if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: output}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(output, ".gitlab-ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// A variable carrying a description stays an object, so "value" and
// "description" keep keys of their own and only the name is left without one.
// The collapsed form carried its name comment and this one did not.
func TestAnExpandedVariableKeepsItsNameComment(t *testing.T) {
	hcl := "variable \"V\" {\n" +
		"  name        = \"V\" # what this sets\n" +
		"  value       = \"x\"\n" +
		"  description = \"d\"\n}\n"

	want := "V: # what this sets"

	if got := parseVariablePipeline(t, hcl); !strings.Contains(got, want) {
		t.Errorf("want %q in:\n%s", want, got)
	}
}
