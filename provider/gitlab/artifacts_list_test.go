// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// "cache" and "artifacts" share an unparse branch that writes one block per
// list entry. A job takes many caches but only one artifacts, so a two-entry
// artifacts list wrote HCL parse refuses.
func TestArtifactsTakesASingleObject(t *testing.T) {
	for _, tc := range []struct {
		name    string
		yml     string
		wantErr error
		wantHCL []string
	}{
		{
			name:    "two entries are refused",
			yml:     "build:\n  script: [make]\n  artifacts:\n    - paths: [a]\n    - paths: [b]\n",
			wantErr: errArtifactsNotAList,
		},
		{
			// "default" goes down the generic writer rather than the job one,
			// which had no cap at all: two blocks went out and parse then
			// refused the file that unparse had just written.
			name:    "two entries under default are refused",
			yml:     "default:\n  artifacts:\n    - paths: [a]\n    - paths: [b]\nbuild:\n  script: [make]\n",
			wantErr: errArtifactsNotAList,
		},
		{
			name:    "two reports are refused",
			yml:     "default:\n  artifacts:\n    reports:\n      - junit: a.xml\n      - junit: b.xml\nbuild:\n  script: [make]\n",
			wantErr: errReportsNotAList,
		},
		{
			name:    "one artifacts under default is written",
			yml:     "default:\n  artifacts:\n    paths: [a]\n    reports:\n      junit: a.xml\nbuild:\n  script: [make]\n",
			wantHCL: []string{"artifacts {", "reports {", `junit = "a.xml"`},
		},
		{
			name:    "a cache list under default keeps repeating",
			yml:     "default:\n  cache:\n    - key: k1\n      paths: [x]\n    - key: k2\n      paths: [y]\nbuild:\n  script: [make]\n",
			wantHCL: []string{`key   = "k1"`, `key   = "k2"`},
		},
		{
			name:    "the object form is written",
			yml:     "build:\n  script: [make]\n  artifacts:\n    paths: [a]\n",
			wantHCL: []string{"artifacts {", `paths = ["a"]`},
		},
		{
			name:    "a single entry list is written",
			yml:     "build:\n  script: [make]\n  artifacts:\n    - paths: [a]\n",
			wantHCL: []string{"artifacts {", `paths = ["a"]`},
		},
		{
			name:    "a cache list keeps repeating",
			yml:     "build:\n  script: [make]\n  cache:\n    - key: k1\n      paths: [x]\n    - key: k2\n      paths: [y]\n",
			wantHCL: []string{`key   = "k1"`, `key   = "k2"`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			in := filepath.Join(tmp, ".gitlab-ci.yml")
			hclDir := filepath.Join(tmp, "hcl")

			if err := os.WriteFile(in, []byte(tc.yml), 0o600); err != nil {
				t.Fatal(err)
			}

			p := New()
			err := p.Unparse(provider.ProviderOps{File: in, OutputDirectory: hclDir})

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("Unparse() error = %v", err)
			}

			hclPath := filepath.Join(hclDir, ".gitlab-ci.hcl")

			got, err := os.ReadFile(hclPath)
			if err != nil {
				t.Fatal(err)
			}

			for _, want := range tc.wantHCL {
				if !strings.Contains(string(got), want) {
					t.Errorf("HCL missing %q:\n%s", want, got)
				}
			}

			// Whatever was written has to be readable again: that is the
			// property the multi-entry list broke.
			if err := p.Parse(provider.ProviderOps{File: hclPath, OutputDirectory: filepath.Join(tmp, "yaml")}); err != nil {
				t.Fatalf("Parse() error = %v\nHCL:\n%s", err, got)
			}
		})
	}
}
