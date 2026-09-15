// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"strings"
	"testing"
)

// The job order comes from the raw key nodes while the jobs map comes from the
// decoder, and the two read a key differently. An alias in key position holds
// the anchor's name, not the text it stands for, so reading it without
// resolving put a name in the order that the map never held.
func TestAliasInJobKeyPositionResolves(t *testing.T) {
	yaml := "on: push\n" +
		"jobs:\n" +
		"  first:\n" +
		"    runs-on: &runner ubuntu-latest\n" +
		"    steps:\n" +
		"      - run: echo first\n" +
		"  *runner :\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - run: echo second\n"

	_, order, err := parseYAMLDocument([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}

	// The decoder resolves the alias to "ubuntu-latest", so the order has to
	// name the same job, not the anchor "runner".
	want := []string{"first", "ubuntu-latest"}

	if len(order) != len(want) {
		t.Fatalf("got order %q, want %q", order, want)
	}

	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("got order %q, want %q", order, want)
		}
	}

	got := unparse(t, yaml)

	for _, name := range []string{"echo first", "echo second"} {
		if !strings.Contains(got, name) {
			t.Errorf("%q is missing from the output:\n%s", name, got)
		}
	}
}

// A job with no name cannot be referred to by needs and has no reference to
// emit, and the HCL side already refuses one on the way back, so the YAML side
// has to say so rather than emit a job whose id is empty.
func TestEmptyJobKeyIsRejected(t *testing.T) {
	yaml := "on: push\n" +
		"jobs:\n" +
		"  build:\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - run: echo hi\n" +
		"  \"\":\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - run: echo unnamed\n"

	err := unparseErr(t, yaml)
	if err == nil {
		t.Fatal("an unnamed job was accepted")
	}

	if !strings.Contains(err.Error(), "non-empty string") {
		t.Errorf("the error does not say the name is the problem: %v", err)
	}
}

// yaml.v3 decodes a mapping holding a non-string key into map[any]any, and a
// null key lands there as a nil that crashed the encoder the validator runs
// the document through. A number or a boolean did not crash but went missing
// just as quietly, so both are refused before any of that.
func TestNonStringMappingKeyIsRejected(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{
			// This one used to be a segmentation fault, not an error.
			name: "null job key",
			yaml: "on: push\njobs:\n  ~:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n",
		},
		{
			name: "null key nested in a step",
			yaml: "on: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n        env:\n          ~: value\n",
		},
		{
			name: "integer job key",
			yaml: "on: push\njobs:\n  123:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n",
		},
		{
			name: "boolean job key",
			yaml: "on: push\njobs:\n  true:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n",
		},
		{
			name: "float key nested in a step",
			yaml: "on: push\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n        env:\n          1.5: value\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := unparseErr(t, tc.yaml)
			if err == nil {
				t.Fatal("a non-string key was accepted")
			}

			if !strings.Contains(err.Error(), "must be a string") {
				t.Errorf("the error does not name the key as the problem: %v", err)
			}
		})
	}
}

// A quoted key that only looks like a number is a string, and a merge key is
// folded away by the decoder before anything else sees it. Neither may be
// caught by the check above.
func TestStringLikeAndMergeKeysStillWork(t *testing.T) {
	yaml := "on: push\n" +
		"jobs:\n" +
		"  \"123\":\n" +
		"    runs-on: &runner ubuntu-latest\n" +
		"    steps:\n" +
		"      - run: echo hi\n" +
		"        env:\n" +
		"          \"1.5\": value\n" +
		"  other:\n" +
		"    <<: {timeout-minutes: 3}\n" +
		"    runs-on: *runner\n" +
		"    steps:\n" +
		"      - run: echo other\n"

	got := unparse(t, yaml)

	if !strings.Contains(got, "timeout_minutes = 3") {
		t.Errorf("the merge key did not survive:\n%s", got)
	}

	// An HCL label cannot start with a digit, so the name is sanitised and
	// the original kept as the id.
	if !strings.Contains(got, `id = "123"`) {
		t.Errorf("the quoted numeric job name did not survive:\n%s", got)
	}

	if !strings.Contains(got, `name  = "1.5"`) {
		t.Errorf("the quoted float key did not survive:\n%s", got)
	}
}

// The decoder folds a merge key inside the jobs mapping into the jobs
// themselves, under keys the source order pass cannot see. Reading the key
// literally put "<<" in the order and lost the merged job, so the order falls
// back to sorted keys when one is present.
func TestMergeKeyInsideJobsKeepsEveryJob(t *testing.T) {
	yaml := "on: push\n" +
		"jobs:\n" +
		"  <<: {merged: {runs-on: ubuntu-latest, steps: [{run: echo merged}]}}\n" +
		"  listed:\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - run: echo listed\n"

	if _, order, err := parseYAMLDocument([]byte(yaml)); err != nil {
		t.Fatal(err)
	} else if len(order) != 0 {
		t.Errorf("want no source order when a merge key is present, got %q", order)
	}

	got := unparse(t, yaml)

	for _, want := range []string{`job "merged"`, `job "listed"`} {
		if !strings.Contains(got, want) {
			t.Errorf("%s is missing from the output:\n%s", want, got)
		}
	}
}
