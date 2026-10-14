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

func compositeAction(name string) string {
	return "name: " + name + "\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: Go\n      run: echo hi\n      shell: bash\n"
}

func actionWithPathHCL(label, filename, actionPath string) string {
	block := "step \"" + label + "_go\" {\n  run = \"echo hi\"\n}\n\naction \"" + label + "\" {\n  filename = \"" + filename + "\"\n"

	if actionPath != "" {
		block += "  path     = \"" + actionPath + "\"\n"
	}

	return block + "  name     = \"" + label + "\"\n\n  description = \"d\"\n\n  runs {\n    using = \"composite\"\n    steps = [step." + label + "_go]\n  }\n}\n"
}

// An action can sit anywhere a repository puts it, so where it sits has to
// survive the conversion. Without "path" every action came back under the
// output directory, which for a repo keeping them in actions-a/ meant one
// parse could not put its own files back.
func TestActionPathDecidesWhereItIsWritten(t *testing.T) {
	t.Run("an action with a path writes under it", func(t *testing.T) {
		tmp := t.TempDir()
		src := filepath.Join(tmp, "in.hcl")
		out := filepath.Join(tmp, "out")

		if err := os.WriteFile(src, []byte(actionWithPathHCL("setup", "setup", "actions-a")), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(filepath.Join(out, "actions-a", "setup", "action.yml")); err != nil {
			t.Errorf("want the action under its path: %v", err)
		}
	})

	// The behaviour every file written before this attribute existed relies
	// on. An absent path is not an empty one.
	t.Run("an action without a path is unchanged", func(t *testing.T) {
		tmp := t.TempDir()
		src := filepath.Join(tmp, "in.hcl")
		out := filepath.Join(tmp, "out")

		if err := os.WriteFile(src, []byte(actionWithPathHCL("setup", "setup", "")), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(filepath.Join(out, "setup", "action.yml")); err != nil {
			t.Errorf("want the action directly under the output directory: %v", err)
		}
	})

	// Refused before this attribute existed: the collision check saw only the
	// basename, so two actions called "setup" were two writes to one path.
	t.Run("two actions share a name under different paths", func(t *testing.T) {
		tmp := t.TempDir()
		src := filepath.Join(tmp, "in.hcl")
		out := filepath.Join(tmp, "out")

		both := actionWithPathHCL("setup_a", "setup", "actions-a") + "\n" + actionWithPathHCL("setup_b", "setup", "actions-b")

		if err := os.WriteFile(src, []byte(both), 0o600); err != nil {
			t.Fatal(err)
		}

		if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
			t.Fatalf("two actions under different paths were refused: %v", err)
		}

		for _, want := range []string{"actions-a", "actions-b"} {
			if _, err := os.Stat(filepath.Join(out, want, "setup", "action.yml")); err != nil {
				t.Errorf("want %s/setup/action.yml: %v", want, err)
			}
		}
	})

	// A path reaches the filesystem from a file that may have arrived in a
	// pull request, so it is fenced exactly as filename is. "a/../../x" only
	// escapes once joined, which is why the join is checked as well as the
	// parts.
	t.Run("a path cannot climb out", func(t *testing.T) {
		for _, bad := range []string{"../escape", "/etc", "a/../../x"} {
			tmp := t.TempDir()
			src := filepath.Join(tmp, "in.hcl")
			out := filepath.Join(tmp, "out")

			if err := os.WriteFile(src, []byte(actionWithPathHCL("x", "setup", bad)), 0o600); err != nil {
				t.Fatal(err)
			}

			err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out})
			if err == nil {
				t.Errorf("path %q was accepted", bad)

				continue
			}

			if !strings.Contains(err.Error(), "stay inside") {
				t.Errorf("path %q failed for another reason: %v", bad, err)
			}

			if entries, _ := os.ReadDir(out); len(entries) != 0 {
				t.Errorf("path %q wrote %d entries before being refused", bad, len(entries))
			}
		}
	})
}

// Unparse records where it found an action, so a repository keeping actions
// in their own folders round-trips in one parse rather than one per folder.
func TestUnparseRecordsWhereAnActionLives(t *testing.T) {
	tmp := t.TempDir()
	hcl := filepath.Join(tmp, "hcl")

	// Each action gets its own step name. Two unparse runs each allocate
	// step labels from a fresh set, so two actions whose steps are both
	// called "Go" produce two step "go" blocks and parse refuses the pair —
	// which is how unparse behaved before this change too.
	for dir, name := range map[string]string{"actions-a/setup": "Setup", "actions-b/deploy": "Deploy"} {
		full := filepath.Join(tmp, filepath.FromSlash(dir))

		if err := os.MkdirAll(full, 0o750); err != nil {
			t.Fatal(err)
		}

		action := "name: " + name + "\ndescription: d\nruns:\n  using: composite\n  steps:\n    - name: " + name + " step\n      run: echo hi\n      shell: bash\n"

		if err := os.WriteFile(filepath.Join(full, "action.yml"), []byte(action), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	for _, folder := range []string{"actions-a", "actions-b"} {
		err := New().Unparse(provider.ProviderOps{
			Root:            tmp,
			Directory:       folder,
			Recursive:       true,
			OutputDirectory: hcl,
		})
		if err != nil {
			t.Fatalf("Unparse(%s) = %v", folder, err)
		}
	}

	// Named for the action, not for where it goes: the folders live inside
	// the file as "path".
	for file, wantPath := range map[string]string{"setup.hcl": `path     = "actions-a"`, "deploy.hcl": `path     = "actions-b"`} {
		content, err := os.ReadFile(filepath.Join(hcl, file))
		if err != nil {
			t.Fatalf("want %s: %v", file, err)
		}

		if !strings.Contains(string(content), wantPath) {
			t.Errorf("%s does not record its path:\n%s", file, content)
		}
	}

	// And back, in one parse, to where they started.
	back := filepath.Join(tmp, "back")

	if err := New().Parse(provider.ProviderOps{Directory: hcl, OutputDirectory: back}); err != nil {
		t.Fatalf("Parse() = %v", err)
	}

	for _, want := range []string{"actions-a/setup/action.yml", "actions-b/deploy/action.yml"} {
		if _, err := os.Stat(filepath.Join(back, filepath.FromSlash(want))); err != nil {
			t.Errorf("want %s back where it came from: %v", want, err)
		}
	}
}

// An action read from outside the root has no portable spelling to record: an
// absolute path names one machine's disk, and committing that would make the
// HCL unreadable anywhere else. It keeps the behaviour it had before.
func TestAnActionOutsideTheRootRecordsNoPath(t *testing.T) {
	outside := t.TempDir()
	root := t.TempDir()

	dir := filepath.Join(outside, "setup")

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "action.yml"), []byte(compositeAction("Setup")), 0o600); err != nil {
		t.Fatal(err)
	}

	hcl := filepath.Join(root, "hcl")

	if err := New().Unparse(provider.ProviderOps{Root: root, Directory: dir, Recursive: true, OutputDirectory: hcl}); err != nil {
		t.Fatalf("Unparse() = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(hcl, "setup.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(content), "path") {
		t.Errorf("an absolute location was written into the HCL:\n%s", content)
	}
}

// An action's path is recorded against the run's root, so parse has to place
// it there. Joined onto a parse output directory of .github/workflows
// instead, the action landed at .github/workflows/actions-a/setup/action.yml:
// converted cleanly, pruned correctly, and never read by GitHub, which is the
// worst shape a failure can take.
func TestAnActionWithAPathGoesToTheRootNotTheOutputDirectory(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "in.hcl")

	if err := os.WriteFile(src, []byte(actionWithPathHCL("act", "setup", "actions-a")), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{Root: tmp, File: src}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "actions-a", "setup", "action.yml")); err != nil {
		t.Errorf("want the action at the root: %v", err)
	}

	buried := filepath.Join(tmp, ".github", "workflows", "actions-a", "setup", "action.yml")

	if _, err := os.Stat(buried); !os.IsNotExist(err) {
		t.Errorf("the action was written where GitHub does not read it: %s", buried)
	}
}

// Naming a destination asks for everything to go there, which is what a
// scratch conversion or a test wants.
func TestAnExplicitOutputDirectoryCollectsActionsToo(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "in.hcl")
	out := filepath.Join(tmp, "out")

	if err := os.WriteFile(src, []byte(actionWithPathHCL("act", "setup", "actions-a")), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: src, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(out, "actions-a", "setup", "action.yml")); err != nil {
		t.Errorf("want the action under the named directory: %v", err)
	}
}
