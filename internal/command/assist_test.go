// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitYAMLDocuments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "single document",
			input: "name: test\non:\n  push:",
			want:  1,
		},
		{
			name:  "two documents",
			input: "name: workflow1\non:\n  push:\n---\nname: workflow2\non:\n  pull_request:",
			want:  2,
		},
		{
			name:  "leading separator ignored",
			input: "---\nname: test",
			want:  1,
		},
		{
			name:  "three documents",
			input: "name: a\n---\nname: b\n---\nname: c",
			want:  3,
		},
		{
			name:  "empty input",
			input: "",
			want:  0,
		},
		{
			name:  "whitespace only",
			input: "   \n\n  ",
			want:  0,
		},
		{
			name:  "separator with trailing whitespace",
			input: "name: a\n---  \nname: b",
			want:  2,
		},
		{
			// A separator is only one at column 0. An indented "---" is the
			// content of whatever block scalar it sits in.
			name:  "indented separator is not one",
			input: "name: a\n  ---\nname: b",
			want:  1,
		},
		{
			name:  "--- inside a run block scalar",
			input: "name: test\njobs:\n  build:\n    steps:\n      - run: |\n          cat <<EOF\n          ---\n          key: value\n          EOF\n",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitYAMLDocuments(tt.input)
			if len(got) != tt.want {
				t.Errorf("splitYAMLDocuments() returned %d documents, want %d\ndocs: %v", len(got), tt.want, got)
			}
		})
	}
}

func TestSplitYAMLDocumentsContent(t *testing.T) {
	input := "name: workflow1\non:\n  push:\n---\nname: workflow2\non:\n  pull_request:"
	docs := splitYAMLDocuments(input)

	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}

	if got := docs[0]; got != "name: workflow1\non:\n  push:\n" {
		t.Errorf("doc[0]:\ngot:  %q\nwant: %q", got, "name: workflow1\non:\n  push:\n")
	}

	if got := docs[1]; got != "name: workflow2\non:\n  pull_request:\n" {
		t.Errorf("doc[1]:\ngot:  %q\nwant: %q", got, "name: workflow2\non:\n  pull_request:\n")
	}
}

func TestSplitHCLBlocksAST(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantBlocks int
	}{
		{
			name: "two blocks",
			input: `step "checkout" {
  name = "Checkout"
}

step "test" {
  name = "Test"
}`,
			wantBlocks: 2,
		},
		{
			name: "block with braces in string",
			input: `step "deploy" {
  run = "echo ${VAR}"
}`,
			wantBlocks: 1,
		},
		{
			name: "nested blocks",
			input: `workflow "pr" {
  on "pull_request" {}
  jobs = [job.test]
}`,
			wantBlocks: 1,
		},
		{
			name:       "invalid HCL falls back to single block",
			input:      "this is not valid HCL {{{",
			wantBlocks: 1,
		},
		{
			name: "top-level attribute",
			input: `variable "os" {
  value = "ubuntu"
}

name = "test"`,
			wantBlocks: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitHCLBlocksAST([]byte(tt.input), "test.hcl")
			if len(got) != tt.wantBlocks {
				t.Errorf("splitHCLBlocksAST() returned %d blocks, want %d\nblocks: %v", len(got), tt.wantBlocks, got)
			}
		})
	}
}

func TestMergeHCLFiles(t *testing.T) {
	dir := t.TempDir()

	file1 := `step "checkout" {
  name = "Checkout"
}

step "test" {
  name = "Test"
}
`
	file2 := `step "checkout" {
  name = "Checkout"
}

step "build" {
  name = "Build"
}
`

	if err := os.WriteFile(filepath.Join(dir, "a.hcl"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "b.hcl"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}

	merged, err := mergeHCLFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	// checkout should appear only once (deduped)
	if count := strings.Count(merged, `step "checkout"`); count != 1 {
		t.Errorf("expected 1 checkout block, got %d\nmerged:\n%s", count, merged)
	}

	// test and build should each appear once
	if !strings.Contains(merged, `step "test"`) {
		t.Error("expected test block in merged output")
	}

	if !strings.Contains(merged, `step "build"`) {
		t.Error("expected build block in merged output")
	}
}

func TestMergeHCLFilesIgnoresNonHCL(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Secret"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "test.hcl"), []byte(`step "a" {}`), 0644); err != nil {
		t.Fatal(err)
	}

	merged, err := mergeHCLFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(merged, "Secret") {
		t.Error("non-HCL content should not appear in merged output")
	}
}

func TestConfirmCost(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "yes", input: "y\n", wantErr: false},
		{name: "YES", input: "YES\n", wantErr: false},
		{name: "no", input: "n\n", wantErr: true},
		{name: "empty", input: "\n", wantErr: true},
		{name: "eof", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := confirmCost(&buf, strings.NewReader(tt.input), "anthropic", "default")

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestResolveAIProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
		wantName string
	}{
		{name: "anthropic explicit", provider: "anthropic", wantErr: true},
		{name: "openai explicit", provider: "openai", wantErr: true},
		{name: "empty defaults to anthropic", provider: "", wantErr: true},
		{name: "unknown", provider: "gemini", wantErr: true},
	}

	// All cases error because no API keys are set in test env.
	// We verify provider resolution logic, not API connectivity.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveAIProvider(tt.provider, "")
			if tt.provider == "gemini" {
				if err == nil || !strings.Contains(err.Error(), "unknown AI provider") {
					t.Errorf("expected unknown provider error, got %v", err)
				}
			} else if err == nil {
				t.Error("expected missing API key error without env var")
			}
		})
	}
}

func TestResolveAIProviderWithKey(t *testing.T) {
	p, err := resolveAIProvider("anthropic", "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Name() != "anthropic" {
		t.Errorf("expected anthropic, got %s", p.Name())
	}

	p, err = resolveAIProvider("openai", "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Name() != "openai" {
		t.Errorf("expected openai, got %s", p.Name())
	}
}

func TestValidateRelativePath(t *testing.T) {
	absPath := "/etc/secrets"
	if runtime.GOOS == "windows" {
		absPath = `C:\Windows\System32`
	}

	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{name: "valid relative", path: "cinzel/assist", wantErr: nil},
		{name: "valid simple", path: "output", wantErr: nil},
		{name: "valid nested", path: "a/b/c", wantErr: nil},
		{name: "absolute path", path: absPath, wantErr: errAbsolutePath},
		{name: "parent traversal", path: "../../../etc", wantErr: errPathTraversal},
		{name: "hidden traversal", path: "foo/../../bar", wantErr: errPathTraversal},
		{name: "current dir", path: ".", wantErr: nil},
		// A name that merely starts with ".." escapes nothing. Testing the
		// first two characters refused these.
		{name: "a name beginning with two dots", path: "..hidden", wantErr: nil},
		{name: "three dots", path: "...x", wantErr: nil},
		{name: "two dots inside the path", path: "a/..hidden/b", wantErr: nil},
		{name: "bare parent", path: "..", wantErr: errPathTraversal},
		{name: "parent then a name", path: "../x", wantErr: errPathTraversal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRelativePath(tt.path)
			if tt.wantErr == nil && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.wantErr != nil && err == nil {
				t.Errorf("expected %v, got nil", tt.wantErr)
			}

			if tt.wantErr != nil && err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestBlockSignature(t *testing.T) {
	tests := []struct {
		block string
		want  string
	}{
		{`step "checkout" {` + "\n  name = \"Checkout\"\n}", `step "checkout"`},
		{`workflow "pr" {` + "\n  name = \"PR\"\n}", `workflow "pr"`},
		{`variable "os" {` + "\n  value = []\n}", `variable "os"`},
		// Every pinned action carries a tag comment. The first line used to be
		// taken outright, so the comment became the signature.
		{"// action tag: v4\n" + `step "checkout" {` + "\n  name = \"Checkout\"\n}", `step "checkout"`},
		{"# a hash comment\n\n" + `job "build" {` + "\n}", `job "build"`},
		{"// nothing but a comment\n", ""},
	}

	for _, tt := range tests {
		got := blockSignature(tt.block)
		if got != tt.want {
			t.Errorf("blockSignature(%q) = %q, want %q", tt.block[:20], got, tt.want)
		}
	}
}

func TestDeduplicateWithExisting(t *testing.T) {
	contextDir := t.TempDir()

	existingSteps := `step "checkout" {
  name = "Checkout"

  uses {
    action  = "actions/checkout"
    version = "abc123"
  }
}

step "tests" {
  name = "Tests"
  run  = "go test ./..."
}
`

	if err := os.WriteFile(filepath.Join(contextDir, "steps.hcl"), []byte(existingSteps), 0644); err != nil {
		t.Fatal(err)
	}

	// Generated output has identical checkout, different tests, and a new step.
	generated := `step "checkout" {
  name = "Checkout"

  uses {
    action  = "actions/checkout"
    version = "abc123"
  }
}

step "tests" {
  name = "Tests"
  run  = "npm test"
}

step "deploy" {
  name = "Deploy"
  run  = "deploy.sh"
}
`

	result, _ := deduplicateWithExisting(generated, contextDir)

	// Identical checkout should be replaced with reference.
	if !strings.Contains(result, `// reuses: step "checkout" from steps.hcl`) {
		t.Errorf("expected reuse comment for checkout\ngot:\n%s", result)
	}

	// Checkout block content should NOT be in output.
	if strings.Contains(result, `action  = "actions/checkout"`) {
		t.Errorf("identical checkout block should be replaced, not kept\ngot:\n%s", result)
	}

	// Different tests should be kept with a note.
	if !strings.Contains(result, `// note: step "tests" also exists in steps.hcl`) {
		t.Errorf("expected note for different tests block\ngot:\n%s", result)
	}

	if !strings.Contains(result, `run  = "npm test"`) {
		t.Errorf("different tests block should be kept\ngot:\n%s", result)
	}

	// New step should be kept as-is.
	if !strings.Contains(result, `step "deploy"`) {
		t.Errorf("new deploy step should be kept\ngot:\n%s", result)
	}
}

// A pinned step carries an "// action tag" comment. Its signature used to come
// out as that comment, which matched nothing, so the block was emitted in full
// instead of as a reference to the one already in context.
func TestDeduplicateMatchesACommentedBlock(t *testing.T) {
	contextDir := t.TempDir()

	block := `// action tag: v4
step "checkout" {
  name = "Checkout"

  uses {
    action  = "actions/checkout"
    version = "abc123"
  }
}
`

	if err := os.WriteFile(filepath.Join(contextDir, "steps.hcl"), []byte(block), 0644); err != nil {
		t.Fatal(err)
	}

	result, _ := deduplicateWithExisting(block, contextDir)

	if !strings.Contains(result, `// reuses: step "checkout" from steps.hcl`) {
		t.Errorf("expected a reuse comment naming the block\ngot:\n%s", result)
	}

	if strings.Contains(result, `action  = "actions/checkout"`) {
		t.Errorf("the identical block should be replaced, not kept\ngot:\n%s", result)
	}
}

func TestDeduplicateWithExistingNoContextDir(t *testing.T) {
	input := `step "checkout" {
  name = "Checkout"
}
`
	result, _ := deduplicateWithExisting(input, "/nonexistent/path")

	if result != input {
		t.Errorf("should return input unchanged for nonexistent dir\ngot: %q", result)
	}
}

// "--refine" with no session named picks the last one, which the length check
// got wrong: any 15-character directory counted, so an unrelated one sorting
// above the real sessions was handed to --refine as the last session.
func TestLatestAssistDirTakesOnlyATimestamp(t *testing.T) {
	for _, tc := range []struct {
		name string
		dirs []string
		want string
	}{
		{
			name: "the newest of several sessions",
			dirs: []string{"20260101-101500", "20260918-090000", "20260305-235959"},
			want: "20260918-090000",
		},
		{
			// 15 characters, and it sorts above every real session, so the
			// length check handed this one to --refine.
			name: "a 15-character decoy is skipped",
			dirs: []string{"20260918-090000", "zzzzzzzz-zzzzzz"},
			want: "20260918-090000",
		},
		{
			// The right length and the right shape, but not a date.
			name: "a malformed timestamp is skipped",
			dirs: []string{"20260918-090000", "99999999-999999"},
			want: "20260918-090000",
		},
		{
			name: "no session at all",
			dirs: []string{"notes", "zzzzzzzz-zzzzzz"},
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()

			for _, dir := range tc.dirs {
				if err := os.Mkdir(filepath.Join(base, dir), 0o750); err != nil {
					t.Fatal(err)
				}
			}

			want := ""
			if tc.want != "" {
				want = filepath.Join(base, tc.want)
			}

			if got := latestAssistDir(base); got != want {
				t.Errorf("latestAssistDir() = %q, want %q", got, want)
			}
		})
	}
}

// The preview of a failed conversion is cut to a byte budget. A cut that lands
// inside a multibyte rune leaves a partial encoding behind, and that reaches
// the terminal as U+FFFD rather than as the character the model wrote.
func TestErrorPreviewIsNotCutMidRune(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
	}{
		{"two-byte rune on the boundary", strings.Repeat("a", maxRawYAMLErrorLen-1) + "é" + "tail"},
		{"four-byte rune on the boundary", strings.Repeat("a", maxRawYAMLErrorLen-2) + "🙂" + "tail"},
		{"rune straddling by one byte", strings.Repeat("a", maxRawYAMLErrorLen-3) + "🙂" + "tail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := truncatePreview(tc.source, maxRawYAMLErrorLen)

			if !utf8.ValidString(got) {
				t.Errorf("preview is not valid UTF-8: %q", got)
			}

			if strings.ContainsRune(got, utf8.RuneError) {
				t.Errorf("preview holds U+FFFD: %q", got)
			}

			content := strings.TrimSuffix(got, truncatedSuffix)
			if content == got {
				t.Fatalf("preview was not marked as truncated: %q", got)
			}

			if len(content) > maxRawYAMLErrorLen {
				t.Errorf("preview content is %d bytes, over the %d budget", len(content), maxRawYAMLErrorLen)
			}
		})
	}
}

// A preview inside the budget is returned whole.
func TestShortErrorPreviewIsUntouched(t *testing.T) {
	const s = "name: Déploiement 🙂\n"

	if got := truncatePreview(s, maxRawYAMLErrorLen); got != s {
		t.Errorf("truncatePreview(%q) = %q, want it unchanged", s, got)
	}
}

// A context directory nobody created is the normal case and says nothing. One
// that exists and cannot be read is different: deduplication silently does not
// happen, so assist repeats blocks the user already has, with no sign why.
func TestUnreadableContextDirIsReported(t *testing.T) {
	// Windows has no Unix permission bits: os.Mkdir's mode only toggles the
	// read-only flag there, which does not stop a directory being listed.
	if runtime.GOOS == "windows" {
		t.Skip("a 0000 directory is still readable on Windows")
	}

	if os.Geteuid() == 0 {
		t.Skip("root reads a 0000 directory regardless of its mode")
	}

	dir := filepath.Join(t.TempDir(), "cinzel")
	if err := os.Mkdir(dir, 0000); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })

	_, warnings := deduplicateWithExisting("step \"x\" {\n  run = \"echo\"\n}\n", dir)

	if len(warnings) == 0 {
		t.Fatal("no warning for a context directory that could not be read")
	}

	if !strings.Contains(warnings[0], dir) {
		t.Errorf("warning does not name the directory: %q", warnings[0])
	}
}

// A context directory that is simply absent is silent.
func TestMissingContextDirIsSilent(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")

	merged := "step \"x\" {\n  run = \"echo\"\n}\n"

	got, warnings := deduplicateWithExisting(merged, missing)

	if len(warnings) != 0 {
		t.Errorf("warnings = %q, want none", warnings)
	}

	if got != merged {
		t.Errorf("merged content changed: %q", got)
	}
}

// A file inside the context directory that cannot be read is a block that will
// not be deduplicated against, which is worth the same warning.
func TestUnreadableContextFileIsReported(t *testing.T) {
	// Windows has no Unix permission bits: a 0000 file reads back as 0666 and
	// opens normally. See TestInitTightensPermissionsOnAnExistingFile.
	if runtime.GOOS == "windows" {
		t.Skip("a 0000 file is still readable on Windows")
	}

	if os.Geteuid() == 0 {
		t.Skip("root reads a 0000 file regardless of its mode")
	}

	dir := t.TempDir()

	blocked := filepath.Join(dir, "blocked.hcl")
	if err := os.WriteFile(blocked, []byte("step \"x\" {\n  run = \"echo\"\n}\n"), 0000); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(blocked, 0600) })

	_, warnings := deduplicateWithExisting("step \"y\" {\n  run = \"echo\"\n}\n", dir)

	if len(warnings) == 0 {
		t.Fatal("no warning for a context file that could not be read")
	}

	if !strings.Contains(warnings[0], "blocked.hcl") {
		t.Errorf("warning does not name the file: %q", warnings[0])
	}
}
