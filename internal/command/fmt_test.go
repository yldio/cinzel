// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

const unformattedHCL = "workflow \"ci\" {\n      name = \"CI\"\n}\n"

// runFmt runs "cinzel fmt" from dir, which is where the relative paths the
// command takes are resolved.
func runFmt(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()

	t.Chdir(dir)

	out := new(bytes.Buffer)
	app := NewWithErrWriter(out, new(bytes.Buffer), "v.test")

	runErr := app.Execute(append([]string{"cinzel", "fmt"}, args...), []provider.Provider{})

	return out.String(), runErr
}

func writeHCL(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

func TestFmtRewritesAndIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cinzel", "workflows.hcl")
	writeHCL(t, path, unformattedHCL)

	out, err := runFmt(t, dir)
	if err != nil {
		t.Fatalf("fmt: %v", err)
	}

	first := readFile(t, path)

	if first == unformattedHCL {
		t.Fatal("fmt left the file unformatted")
	}

	if !strings.Contains(out, "1 formatted") {
		t.Fatalf("summary did not count the rewrite: %q", out)
	}

	if _, err := runFmt(t, dir); err != nil {
		t.Fatalf("second fmt: %v", err)
	}

	if second := readFile(t, path); second != first {
		t.Fatalf("fmt is not idempotent:\nfirst:  %q\nsecond: %q", first, second)
	}
}

// --check reports and exits non-zero, and above all writes nothing: a CI step
// calling it on a drifted file must fail rather than quietly fix it.
func TestFmtCheckReportsWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cinzel", "workflows.hcl")
	writeHCL(t, path, unformattedHCL)

	out, err := runFmt(t, dir, "--check")
	if !errors.Is(err, errUnformatted) {
		t.Fatalf("want errUnformatted, got %v", err)
	}

	if !strings.Contains(out, "workflows.hcl") {
		t.Fatalf("the unformatted file was not named: %q", out)
	}

	if got := readFile(t, path); got != unformattedHCL {
		t.Fatalf("--check wrote to the file: %q", got)
	}
}

func TestFmtCheckPassesOnFormattedFiles(t *testing.T) {
	dir := t.TempDir()
	writeHCL(t, filepath.Join(dir, "cinzel", "workflows.hcl"), "workflow \"ci\" {\n  name = \"CI\"\n}\n")

	if _, err := runFmt(t, dir, "--check"); err != nil {
		t.Fatalf("fmt --check on a formatted file: %v", err)
	}
}

// hclwrite.Format works on tokens, so it reformats HCL it cannot parse as
// readily as HCL it can. Writing that back would leave a broken file laid out
// neatly enough to look deliberate.
func TestFmtLeavesUnparseableHCLAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cinzel", "broken.hcl")
	broken := "workflow \"ci\" {\n      name =\n"
	writeHCL(t, path, broken)

	if _, err := runFmt(t, dir); err == nil {
		t.Fatal("fmt accepted unparseable HCL")
	}

	if got := readFile(t, path); got != broken {
		t.Fatalf("fmt rewrote unparseable HCL: %q", got)
	}
}

func TestFmtReachesNestedFilesOnlyWhenRecursive(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "cinzel", "github", "workflows.hcl")
	writeHCL(t, nested, unformattedHCL)

	if _, err := runFmt(t, dir); err != nil {
		t.Fatalf("fmt: %v", err)
	}

	if got := readFile(t, nested); got != unformattedHCL {
		t.Fatalf("fmt descended without --recursive: %q", got)
	}

	if _, err := runFmt(t, dir, "--recursive"); err != nil {
		t.Fatalf("fmt --recursive: %v", err)
	}

	if got := readFile(t, nested); got == unformattedHCL {
		t.Fatal("fmt --recursive left the nested file unformatted")
	}
}

func TestFmtRefusesAPathOutsideTheProject(t *testing.T) {
	dir := t.TempDir()
	writeHCL(t, filepath.Join(dir, "cinzel", "workflows.hcl"), unformattedHCL)

	if _, err := runFmt(t, dir, "--directory", ".."); err == nil {
		t.Fatal("fmt accepted a directory outside the project")
	}
}
