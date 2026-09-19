// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"strings"
	"testing"
)

// needsHCL returns two job blocks, the second waiting on the first, and the
// workflow bodies given.
func needsHCL(workflows string) string {
	return "step \"s\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"alpha\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n  steps = [step.s]\n}\n\n" +
		"job \"beta\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n" +
		"  depends_on = [job.alpha]\n\n  steps = [step.s]\n}\n\n" + workflows
}

// GitHub reads "needs" within one file, and the jobs a workflow writes are the
// ones it lists. A depends_on reaching a job block another workflow lists wrote
// a name nothing in the file answers to: actionlint reports `job "beta" needs
// job "alpha" which does not exist in this workflow`. The check ran over every
// job block at once, so the reference went through.
func TestAJobCannotWaitOnAJobAnotherWorkflowWrites(t *testing.T) {
	hcl := needsHCL(
		"workflow \"one\" {\n  filename = \"one\"\n\n  on \"push\" {}\n\n  jobs = [job.beta]\n}\n\n" +
			"workflow \"two\" {\n  filename = \"two\"\n\n  on \"push\" {}\n\n  jobs = [job.alpha]\n}\n")

	err := parseHCLString(t, hcl)

	if err == nil {
		t.Fatal("a job waited on one no workflow in the file writes beside it")
	}

	if !errors.Is(err, errNeedsOutsideWorkflow) {
		t.Fatalf("the error does not say the job is outside the workflow: %v", err)
	}
}

// A workflow that lists both jobs keeps the wait it wrote.
func TestAJobWaitsOnAJobTheSameWorkflowWrites(t *testing.T) {
	hcl := needsHCL(
		"workflow \"wf\" {\n  filename = \"wf\"\n\n  on \"push\" {}\n\n  jobs = [job.alpha, job.beta]\n}\n")

	got := parseHCLToYAML(t, hcl)

	if !strings.Contains(got, "- alpha") {
		t.Fatalf("the wait the job wrote is gone:\n%s", got)
	}
}
