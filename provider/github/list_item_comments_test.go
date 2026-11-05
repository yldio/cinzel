// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// A comment on one entry of a list is a note on that entry, not on the
// attribute holding it. Neither direction had anywhere to put one: the list
// went through as a plain slice either way, so "- edited # re-check the title"
// came back as a bare "edited" and the note was gone.
func TestACommentOnAListEntrySurvivesUnparse(t *testing.T) {
	const yaml = "name: CI\non:\n  pull_request:\n    types:\n      - opened\n" +
		"      # the title may have changed\n      - edited # re-check it\n      - reopened\n" +
		"jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"

	hcl := unparseWorkflow(t, yaml)

	for _, want := range []string{"# the title may have changed", "\"edited\", # re-check it"} {
		if !strings.Contains(hcl, want) {
			t.Errorf("want %q in the HCL:\n%s", want, hcl)
		}
	}
}

func TestACommentOnAListEntrySurvivesParse(t *testing.T) {
	const src = "workflow \"w\" {\n  filename = \"w\"\n\n  name = \"CI\"\n\n" +
		"  on \"pull_request\" {\n    types = [\n      \"opened\",\n" +
		"      # the title may have changed\n      \"edited\", # re-check it\n" +
		"      \"reopened\",\n    ]\n  }\n\n  jobs = [job.b]\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  steps = [step.hi]\n}\n\n" +
		"step \"hi\" {\n  run = \"echo hi\"\n}\n"

	yaml := parsedYAML(t, src)

	for _, want := range []string{"# the title may have changed\n      - edited # re-check it"} {
		if !strings.Contains(yaml, want) {
			t.Errorf("want %q in the YAML:\n%s", want, yaml)
		}
	}
}

// The comment has to come back where it was written after a full trip out and
// back, which is the thing an author notices: a roundtrip that moves it onto
// the neighbouring entry is as wrong as one that drops it.
func TestAListEntryCommentIsStableOverARoundtrip(t *testing.T) {
	const src = "workflow \"w\" {\n  filename = \"w\"\n\n  name = \"CI\"\n\n" +
		"  on \"pull_request\" {\n    types = [\n      \"opened\",\n" +
		"      # the title may have changed\n      \"edited\", # re-check it\n" +
		"      \"reopened\",\n    ]\n  }\n\n  jobs = [job.b]\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  steps = [step.hi]\n}\n\n" +
		"step \"hi\" {\n  run = \"echo hi\"\n}\n"

	back := unparseWorkflow(t, parsedYAML(t, src))

	if !strings.Contains(back, "# the title may have changed\n      \"edited\", # re-check it") {
		t.Errorf("the comment did not come back where it was written:\n%s", back)
	}
}

// A list with nothing written on its entries keeps the one-line form. The
// multi-line writer is for the entries that need a line each, and turning
// every list in every file into a column would rewrite output that was right.
func TestAnUncommentedListStaysOnOneLine(t *testing.T) {
	const yaml = "name: CI\non:\n  pull_request:\n    types:\n      - opened\n      - edited\n" +
		"jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n"

	if hcl := unparseWorkflow(t, yaml); !strings.Contains(hcl, `types = ["opened", "edited"]`) {
		t.Errorf("want the list on one line:\n%s", hcl)
	}
}
