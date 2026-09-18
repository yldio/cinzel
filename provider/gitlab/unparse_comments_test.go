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

// unparsePipeline runs the YAML through the provider and hands back the HCL it
// wrote, which is where a comment either survived or did not.
func unparsePipeline(t *testing.T, src string) string {
	t.Helper()

	dir := t.TempDir()
	in := filepath.Join(dir, ".gitlab-ci.yml")
	out := filepath.Join(dir, "hcl")

	if err := os.WriteFile(in, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatalf("Unparse() error = %v\nYAML:\n%s", err, src)
	}

	content, err := os.ReadFile(filepath.Join(out, ".gitlab-ci.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// unparseJobYAML returns a pipeline whose build job holds the given lines,
// which is where each case writes its comment.
func unparseJobYAML(body string) string {
	if body == "" {
		body = "  script:\n    - make\n"
	}

	return "stages:\n  - build\n\nbuild:\n  stage: build\n" + body
}

func TestCommentsSurviveUnparse(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "above a top-level attribute",
			yaml: "# the stages we run\nstages:\n  - build\n\nbuild:\n  script:\n    - make\n",
			want: "# the stages we run\nstages",
		},
		{
			name: "trailing a top-level attribute",
			yaml: "stages: # in order\n  - build\n\nbuild:\n  script:\n    - make\n",
			want: "# in order",
		},
		{
			name: "above a job",
			yaml: "stages:\n  - build\n\n# the build job\nbuild:\n  script:\n    - make\n",
			want: "# the build job\njob \"build\"",
		},
		{
			name: "above a job attribute",
			yaml: unparseJobYAML("  # what it runs\n  script:\n    - make\n"),
			want: "# what it runs\n  script",
		},
		{
			name: "trailing a job attribute",
			yaml: unparseJobYAML("  script: # one command\n    - make\n"),
			want: "# one command",
		},
		{
			name: "closing a job",
			yaml: unparseJobYAML("  script:\n    - make\n  # that is the whole job\n"),
			want: "# that is the whole job\n}",
		},
		{
			name: "above a block",
			yaml: unparseJobYAML("  script:\n    - make\n  # keep the binary\n  artifacts:\n    paths:\n      - bin/\n"),
			want: "# keep the binary\n  artifacts",
		},
		{
			name: "closing a block",
			yaml: unparseJobYAML("  script:\n    - make\n  artifacts:\n    paths:\n      - bin/\n    # nothing else\n"),
			want: "# nothing else\n  }",
		},
		{
			name: "above one entry of a list of blocks",
			yaml: unparseJobYAML("  script:\n    - make\n  rules:\n    # only on main\n    - if: \"x\"\n    - when: never\n"),
			want: "# only on main\n  rule",
		},
		{
			name: "inside a variable written as a mapping",
			yaml: "variables:\n  MODE:\n    value: fast # the default\n    description: how\n\nbuild:\n  script:\n    - make\n",
			want: "\"fast\" # the default",
		},
		{
			name: "on a variable written as a scalar",
			yaml: "variables:\n  # what it holds\n  REGISTRY: example.com # the registry\n\nbuild:\n  script:\n    - make\n",
			want: "\"example.com\" # the registry",
		},
		{
			name: "above a workflow rule",
			yaml: "workflow:\n  rules:\n    # always\n    - when: always\n\nbuild:\n  script:\n    - make\n",
			want: "# always\n  rule",
		},
		{
			name: "above an include",
			yaml: "include:\n  # from elsewhere\n  - local: other.yml\n\nbuild:\n  script:\n    - make\n",
			want: "# from elsewhere\ninclude",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := unparsePipeline(t, tc.yaml)

			if !strings.Contains(got, tc.want) {
				t.Errorf("want %q in output, got:\n%s", tc.want, got)
			}
		})
	}
}

// A comment carried to the wrong key reads as being about something its author
// never wrote it about, which is worse than losing it.
func TestUnparseCommentsLandOnlyWhereTheyWereWritten(t *testing.T) {
	got := unparsePipeline(t, unparseJobYAML("  # about the script\n  script:\n    - make\n  image: alpine\n"))

	if !strings.Contains(got, "# about the script\n  script") {
		t.Errorf("want the comment above script, got:\n%s", got)
	}

	if strings.Contains(got, "# about the script\n  image") {
		t.Errorf("want the comment off image, got:\n%s", got)
	}
}

// A comment that went out has to come back on the same key, or a roundtrip
// quietly moves it every time it runs.
func TestACommentSurvivesAGitLabRoundtrip(t *testing.T) {
	const src = "stages:\n  - build\n\n# the build job\nbuild:\n  # what it runs\n  script: # one command\n    - make\n  stage: build\n  # that is the whole job\n"

	hclSrc := unparsePipeline(t, src)
	got := parsePipeline(t, hclSrc)

	for _, want := range []string{
		"# the build job",
		"# what it runs",
		"# one command",
		"# that is the whole job",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q after a roundtrip, got:\n%s", want, got)
		}
	}
}
