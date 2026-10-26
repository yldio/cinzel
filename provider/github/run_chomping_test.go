// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A script is written as a heredoc, which ends on its closing marker's own
// line and so always holds exactly one trailing newline. That is "run: |".
// "|-" and "|+" mean something a heredoc cannot spell, so both come back as
// "|" and the conversion says so rather than changing the file in silence.
func TestAChompedRunIsReported(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  string
		want string
	}{
		{
			name: "strip",
			run:  "      - run: |-\n          echo one\n          echo two\n",
			want: "rewriting the script's '|-' as '|'",
		},
		{
			// The blank lines the "+" keeps are the last thing in the file:
			// anything after them at a shallower indent ends the block and
			// the keep has nothing to hold.
			name: "keep",
			run:  "      - run: |+\n          echo one\n          echo two\n\n",
			want: "rewriting the script's '|+' as '|'",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			yaml := "on:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n" + tc.run

			_, warnings := unparseWithWarnings(t, yaml)

			if !strings.Contains(warnings, tc.want) {
				t.Errorf("want %q, got:\n%s", tc.want, warnings)
			}

			if !strings.Contains(warnings, "jobs.build.steps[0].run") {
				t.Errorf("the report does not name the script, got:\n%s", warnings)
			}

			// stepFingerprint builds a step twice to measure it. Reporting
			// from there would say the same thing three times over.
			if got := strings.Count(warnings, "rewriting the script"); got != 1 {
				t.Errorf("want one report, got %d:\n%s", got, warnings)
			}
		})
	}
}

// "run: |" is what a heredoc holds, so it makes the trip unchanged and there
// is nothing to report. Nor is a single-line script, which is not a block.
func TestAnUnchompedRunIsNotReported(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  string
	}{
		{name: "clip", run: "      - run: |\n          echo one\n          echo two\n"},
		{name: "one line", run: "      - run: echo one\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			yaml := "on:\n  push: {}\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n" + tc.run

			_, warnings := unparseWithWarnings(t, yaml)

			if strings.Contains(warnings, "rewriting the script") {
				t.Errorf("nothing was rewritten, but it was reported:\n%s", warnings)
			}
		})
	}
}

// Only "run" becomes a heredoc. Every other multi-line string is written as a
// quoted HCL string that holds the value exactly, so its block style survives
// and none of them is reported.
func TestABlockStyleOutsideRunSurvives(t *testing.T) {
	yaml := "on:\n  push: {}\n" +
		"env:\n  TOP: |-\n    env one\n    env two\n" +
		"jobs:\n  build:\n    runs-on: ubuntu-latest\n" +
		"    env:\n      JOB: |-\n        job one\n        job two\n" +
		"    steps:\n      - run: echo hi\n"

	hcl, warnings := unparseWithWarnings(t, yaml)

	if strings.Contains(warnings, "rewriting the script") {
		t.Errorf("a value outside run was reported:\n%s", warnings)
	}

	// The quoted form holds the value with no trailing newline, which is what
	// "|-" means. A heredoc could not have written it.
	for _, want := range []string{`"env one\nenv two"`, `"job one\njob two"`} {
		if !strings.Contains(hcl, want) {
			t.Errorf("want %q in:\n%s", want, hcl)
		}
	}
}
