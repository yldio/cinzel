// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A comment above a job says why the job is there, which is the one thing the
// keys below it cannot. Unparse used to drop it, so the note survived every
// hand edit of the YAML and died the first time the file went through cinzel.
func TestJobCommentsSurviveUnparse(t *testing.T) {
	for _, tc := range []struct {
		name string
		jobs string
		want []string
		gone []string
	}{
		{
			name: "one line above the job",
			jobs: "  # only runs on main\n" + commentJob("check"),
			want: []string{"# only runs on main\njob \"check\""},
		},
		{
			name: "two lines above the job",
			jobs: "  # first\n  # second\n" + commentJob("check"),
			want: []string{"# first\n# second\njob \"check\""},
		},
		{
			name: "a job with no comment is untouched",
			jobs: commentJob("check"),
			gone: []string{"#"},
		},
		{
			name: "only the commented job carries one",
			jobs: "  # about first\n" + commentJob("first") + commentJob("second"),
			want: []string{"# about first\njob \"first\""},
			gone: []string{"# about first\njob \"second\""},
		},
		{
			// The key is sanitized to reach the block label, so the comment
			// has to be found under the key rather than under the label.
			name: "a job whose key is not its label",
			jobs: "  # about the dashed one\n" + commentJob("check-something"),
			want: []string{"# about the dashed one\njob \"check_something\""},
		},
		{
			// A bare "#" carries no text. Written back as "# " it would leave
			// trailing whitespace in generated output.
			name: "an empty comment line",
			jobs: "  #\n" + commentJob("check"),
			want: []string{"#\njob \"check\""},
			gone: []string{"# \n"},
		},
		{
			// A comment is prose. Its spacing is the author's, so a "#" tight
			// against the text stays tight rather than being opened up.
			name: "no space after the hash",
			jobs: "  #no space\n" + commentJob("check"),
			want: []string{"#no space\njob \"check\""},
			gone: []string{"# no space"},
		},
		{
			// A second "#" used to be split off as text and re-prefixed,
			// turning "##" into "# #" and inventing a space mid-comment.
			name: "a double hash and wide spacing",
			jobs: "  ##  banner   style\n" + commentJob("check"),
			want: []string{"##  banner   style\njob \"check\""},
			gone: []string{"# #"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := unparse(t, "name: ci\non:\n  push: {}\njobs:\n"+tc.jobs)

			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("want %q in:\n%s", want, got)
				}
			}

			for _, gone := range tc.gone {
				if strings.Contains(got, gone) {
					t.Errorf("want %q absent from:\n%s", gone, got)
				}
			}
		})
	}
}

// The HCL a comment is written into still has to parse. A comment line that
// reached the file without its "#" would be read as an expression.
func TestACommentedJobStillParses(t *testing.T) {
	got := unparse(t, "name: ci\non:\n  push: {}\njobs:\n"+
		"  # why this job is here\n  # and a second line\n"+commentJob("check"))

	parseWorkflow(t, got)
}

func commentJob(key string) string {
	return "  " + key + ":\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"
}
