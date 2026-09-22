// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
	githubprovider "github.com/yldio/cinzel/provider/github"
)

// Assist writes its session into cinzel/assist/{timestamp}/, under the very
// directory it deduplicates against, so the two are read together. A generated
// block keeping a label the context already holds is then two blocks with one
// label, which parse refuses:
//
//	two step blocks share a label: 'checkout' is declared at ...
//
// The generated one takes a free label instead, and the references in the job
// beside it follow. Dropping it would also parse, and would bind the job to the
// context's block: a workflow that converts cleanly and is not the one the
// generator wrote.
func TestAGeneratedBlockKeepsItsOwnBodyUnderAFreeLabel(t *testing.T) {
	const generated = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

job "build" {
  runs_on { runners = "ubuntu-latest" }
  steps = [step.checkout]
}

workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`

	const existing = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v3"
  }
}
`

	contextDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(contextDir, "steps.hcl"), []byte(existing), 0600); err != nil {
		t.Fatal(err)
	}

	merged, _ := deduplicateWithExisting(generated, contextDir)

	// Parsed the way they are read: the session folder inside the context.
	sessionDir := filepath.Join(contextDir, "assist", "20260922-120000")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(sessionDir, "assist.hcl"), []byte(merged), 0600); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()

	err := githubprovider.New().Parse(provider.ProviderOps{
		Directory:       contextDir,
		OutputDirectory: outDir,
		Recursive:       true,
	})
	if err != nil {
		t.Fatalf("the merged HCL does not parse: %v\n%s", err, merged)
	}

	// The generated body, not the context's. A rename that lost the reference
	// would leave the job on the context's block, which parses just as well and
	// is the wrong workflow, so the version is what tells those two apart.
	out, err := os.ReadFile(filepath.Join(outDir, "ci.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	if got := string(out); !strings.Contains(got, "actions/checkout@v4") {
		t.Errorf("the job does not use the generated block:\n%s", got)
	}
}

// The other branch of the same comparison: an identical block becomes a
// "// reuses:" comment, and the reference then resolves to the context's block,
// which is why nothing is renamed there.
func TestAnIdenticalBlockIsLeftToTheContext(t *testing.T) {
	const step = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}`

	generated := step + `

job "build" {
  runs_on { runners = "ubuntu-latest" }
  steps = [step.checkout]
}

workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`

	contextDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(contextDir, "steps.hcl"), []byte(step+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	merged, _ := deduplicateWithExisting(generated, contextDir)

	if strings.Contains(merged, "checkout_2") {
		t.Errorf("an identical block was renamed, leaving two of it:\n%s", merged)
	}

	sessionDir := filepath.Join(contextDir, "assist", "20260922-120000")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(sessionDir, "assist.hcl"), []byte(merged), 0600); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()

	err := githubprovider.New().Parse(provider.ProviderOps{
		Directory:       contextDir,
		OutputDirectory: outDir,
		Recursive:       true,
	})
	if err != nil {
		t.Fatalf("the merged HCL does not parse: %v\n%s", err, merged)
	}
}

// The rename has to clear this run's own blocks too, not only the context's. A
// free label found against what had been kept so far landed on the label of a
// generated block further down the file, which is the same clash with the names
// swapped, and the job then referenced one block twice.
func TestARenameClearsTheBlocksItHasNotReachedYet(t *testing.T) {
	const generated = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "checkout_2" {
  uses {
    action  = "actions/setup-go"
    version = "v5"
  }
}

job "build" {
  runs_on { runners = "ubuntu-latest" }
  steps = [step.checkout, step.checkout_2]
}

workflow "ci" {
  filename = "ci"
  on "push" {}
  jobs = [job.build]
}
`

	const existing = `step "checkout" {
  uses {
    action  = "actions/checkout"
    version = "v3"
  }
}
`

	contextDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(contextDir, "steps.hcl"), []byte(existing), 0600); err != nil {
		t.Fatal(err)
	}

	merged, _ := deduplicateWithExisting(generated, contextDir)

	sessionDir := filepath.Join(contextDir, "assist", "20260922-120000")

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(sessionDir, "assist.hcl"), []byte(merged), 0600); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()

	err := githubprovider.New().Parse(provider.ProviderOps{
		Directory:       contextDir,
		OutputDirectory: outDir,
		Recursive:       true,
	})
	if err != nil {
		t.Fatalf("the merged HCL does not parse: %v\n%s", err, merged)
	}

	// Both blocks reached the workflow: one renamed, one left where it was. A
	// rename onto an occupied label loses whichever it landed on, and the
	// remaining step list names the survivor twice.
	out, err := os.ReadFile(filepath.Join(outDir, "ci.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	got := string(out)

	for _, want := range []string{"actions/checkout@v4", "actions/setup-go@v5"} {
		if !strings.Contains(got, want) {
			t.Errorf("%s is missing from:\n%s", want, got)
		}
	}
}
