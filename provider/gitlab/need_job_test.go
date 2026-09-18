// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"testing"
)

// A need block is one entry of the YAML "needs" list, and an entry names a
// single job. A longer list used to be dropped without a word, so with
// "pipeline" or "project" also set the file was written and the command exited
// 0 with the jobs it was supposed to wait for gone.
func TestANeedBlockNamesOneJob(t *testing.T) {
	const jobs = "job \"a\" {\n  script = [\"echo a\"]\n  stage  = \"build\"\n}\n\n" +
		"job \"b\" {\n  script = [\"echo b\"]\n  stage  = \"build\"\n}\n\n"

	for _, tc := range []struct {
		name    string
		need    string
		wantErr bool
	}{
		{
			name: "one reference",
			need: "job = job.a",
		},
		{
			name:    "two references",
			need:    "job = [job.a, job.b]",
			wantErr: true,
		},
		{
			name:    "the same reference twice",
			need:    "job = [job.a, job.a]",
			wantErr: true,
		},
		{
			// The case that exited 0: "pipeline" kept the entry valid
			// downstream, so nothing reported the jobs that went missing.
			name:    "two references beside a pipeline",
			need:    "job      = [job.a, job.b]\n    pipeline = \"1234\"",
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hcl := jobs + "job \"c\" {\n  script = [\"echo c\"]\n  stage  = \"test\"\n\n  need {\n    " +
				tc.need + "\n  }\n}\n"

			err := parseHCLString(t, hcl)

			if !tc.wantErr {
				if err != nil {
					t.Fatalf("want no error, got %v", err)
				}

				return
			}

			if !errors.Is(err, errNeedJobNotSingle) {
				t.Fatalf("want %v, got %v", errNeedJobNotSingle, err)
			}
		})
	}
}
