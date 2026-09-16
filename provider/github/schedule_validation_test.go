// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// The schedule check only ran when the event was a map, but the parse path
// denormalizes it to a list first, so a nonsense expression went to YAML and was
// only refused on the way back.
func TestParseValidatesSchedule(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cron    string
		wantErr string
	}{
		{
			name:    "a field out of range",
			cron:    "99 99 99 99 99",
			wantErr: "out of range",
		},
		{
			name:    "the wrong number of fields",
			cron:    "0 9 *",
			wantErr: "must have 5 fields",
		},
		{
			name: "day names are accepted",
			cron: "0 9 * * MON-FRI",
		},
		{
			name: "day-of-week 7 is sunday",
			cron: "0 9 * * 7",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl := "workflow \"in\" {\n  filename = \"in\"\n\n  on \"schedule\" {\n    cron = [\"" + tc.cron + "\"]\n  }\n\n  jobs = [job.a]\n}\n\n" +
				"job \"a\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n  steps = [step.s]\n}\n\n" +
				"step \"s\" {\n  run = \"echo hi\"\n}\n"

			dir := t.TempDir()
			path := filepath.Join(dir, "in.hcl")

			if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
				t.Fatal(err)
			}

			err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("want an error, got nil")
			}

			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want %q in %q", tc.wantErr, err)
			}

			if !strings.Contains(err.Error(), "on.schedule") {
				t.Errorf("want the schedule path in %q", err)
			}
		})
	}
}
