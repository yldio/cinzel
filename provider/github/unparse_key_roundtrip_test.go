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

// unparseJob writes a workflow holding body as the job's contents and returns
// whatever the unparse refused. The body carries its own runs-on, because the
// cases below need a mapping one.
func unparseJob(t *testing.T, body string) (error, string) {
	t.Helper()

	tmp := t.TempDir()
	src := filepath.Join(tmp, "w.yaml")

	content := "on:\n  push:\njobs:\n  build:\n" + body + "    steps:\n      - run: x\n"

	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	err := New().Unparse(provider.ProviderOps{File: src, OutputDirectory: filepath.Join(tmp, "u")})

	out, _ := os.ReadFile(filepath.Join(tmp, "u", "w.hcl"))

	return err, string(out)
}

// A service, a mapping runs-on, a strategy option and a matrix axis all become
// bare HCL identifiers the same way a nested block key does, and none of them
// was checked. An underscore came back as a dash, silently, and a dot produced
// HCL that does not parse at all.
func TestUnparseKeysThatCannotRoundtrip(t *testing.T) {
	const runsOn = "    runs-on: ubuntu-latest\n"

	for _, tc := range []struct {
		name string
		body string
		key  string
		want string
	}{
		{
			name: "a service key holding an underscore",
			body: runsOn + "    services:\n      db:\n        image: postgres\n        my_key: v\n",
			key:  "my_key",
			want: "read back as a dash",
		},
		{
			name: "a service key holding a dot",
			body: runsOn + "    services:\n      db:\n        image: postgres\n        \"a.b\": v\n",
			key:  "a.b",
			want: "not a valid identifier",
		},
		{
			name: "a runs-on key holding an underscore",
			body: "",
			key:  "my_key",
			want: "read back as a dash",
		},
		{
			name: "a runs-on key holding a dot",
			body: "",
			key:  "a.b",
			want: "not a valid identifier",
		},
		{
			name: "a strategy option holding an underscore",
			body: runsOn + "    strategy:\n      my_opt: true\n      matrix:\n        os: [a, b]\n",
			key:  "my_opt",
			want: "read back as a dash",
		},
		{
			name: "a strategy option holding a dot",
			body: runsOn + "    strategy:\n      \"a.b\": true\n      matrix:\n        os: [a, b]\n",
			key:  "a.b",
			want: "not a valid identifier",
		},
		{
			name: "a matrix axis holding a dot",
			body: runsOn + "    strategy:\n      matrix:\n        \"a.b\": one\n",
			key:  "a.b",
			want: "not a valid identifier",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.body

			if body == "" {
				body = "    runs-on:\n      group: g\n      \"" + tc.key + "\": v\n"
			}

			err, out := unparseJob(t, body)

			if err == nil {
				t.Fatalf("expected an error for key %q, got:\n%s", tc.key, out)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error does not name the reason, got: %v", err)
			}

			if !strings.Contains(err.Error(), tc.key) {
				t.Errorf("error does not name the key, got: %v", err)
			}
		})
	}
}

// A matrix axis is written as the author spelled it, because the
// "${{ matrix.X }}" references beside it are copied through unchanged. So an
// underscore there is not the corruption it is anywhere else, and refusing it
// would turn away a workflow GitHub runs.
func TestAMatrixAxisKeepsItsUnderscore(t *testing.T) {
	body := "    runs-on: ubuntu-latest\n    strategy:\n      matrix:\n        go_version: \"1.21\"\n"

	err, out := unparseJob(t, body)

	if err != nil {
		t.Fatalf("expected the axis to be written, got %v", err)
	}

	if !strings.Contains(out, "go_version") {
		t.Errorf("lost the axis name:\n%s", out)
	}
}

// The hyphenated keys the GitHub schema actually uses here have to keep going
// through the same paths the guard now sits on.
func TestUnparseKeysThatDoRoundtrip(t *testing.T) {
	body := "    runs-on:\n      group: g\n      labels: [x]\n" +
		"    services:\n      db:\n        image: postgres\n        credentials:\n          username: u\n" +
		"    strategy:\n      fail-fast: false\n      max-parallel: 2\n      matrix:\n        os: [a, b]\n"

	err, out := unparseJob(t, body)

	if err != nil {
		t.Fatalf("expected the job to be written, got %v", err)
	}

	for _, want := range []string{"fail_fast", "max_parallel", "group", "labels"} {
		if !strings.Contains(out, want) {
			t.Errorf("lost %q:\n%s", want, out)
		}
	}
}
