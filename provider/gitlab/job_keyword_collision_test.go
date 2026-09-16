// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"strings"
	"testing"
)

// The job loop writes into the same map the reserved top-level keys live in.
// A job named after one of them used to overwrite it and exit 0, so a job
// called "stages" took the stage list out of the file.
func TestAJobCannotBeNamedAfterAPipelineKeyword(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want bool
	}{
		{
			name: "stages",
			hcl:  "stages = [\"build\"]\n\n" + jobBlock("stages"),
			want: true,
		},
		{
			name: "variables",
			hcl:  "variable \"A\" {\n  name  = \"A\"\n  value = \"1\"\n}\n\n" + jobBlock("variables"),
			want: true,
		},
		{
			name: "include",
			hcl:  "include {\n  local = \"a.yml\"\n}\n\n" + jobBlock("include"),
			want: true,
		},
		{
			name: "default",
			hcl:  "default {\n  image = \"alpine\"\n}\n\n" + jobBlock("default"),
			want: true,
		},
		{
			name: "the name reached through id",
			hcl:  "stages = [\"build\"]\n\njob \"renamed\" {\n  id     = \"stages\"\n  script = [\"echo x\"]\n  stage  = \"build\"\n}\n",
			want: true,
		},
		{
			name: "an ordinary name",
			hcl:  "stages = [\"build\"]\n\n" + jobBlock("build"),
			want: false,
		},
		{
			name: "a keyword that is not in this pipeline",
			hcl:  jobBlock("stages"),
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseHCLString(t, tc.hcl)

			if !tc.want {
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}

				return
			}

			if !errors.Is(err, errJobNamedAfterKeyword) {
				t.Fatalf("want %v, got %v", errJobNamedAfterKeyword, err)
			}

			// The message has to name the key, or there is nothing to rename.
			if !strings.Contains(err.Error(), "'"+tc.name+"'") && !strings.Contains(err.Error(), "'stages'") {
				t.Errorf("want the colliding name in %q", err)
			}
		})
	}
}

func jobBlock(name string) string {
	return "job \"" + name + "\" {\n  script = [\"echo x\"]\n  stage  = \"build\"\n}\n"
}
