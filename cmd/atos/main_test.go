// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	atosFlag "github.com/yldio/atos/service/flag"
	"github.com/yldio/atos/service/writer"
)

func TestParseHcl(t *testing.T) {
	type ParserTest struct {
		name   string
		have   []byte
		expect []byte
	}

	var have_1 = []byte(`workflow "workflow_1" {
  filename = "dummy-file.yaml"
  on {
    event "push" {
      branches = ["main"]
      tags = ["v2"]
    }
  }
  on {
    activity "label" {
      types = ["created"]
    } 
  }
  jobs = [job.job_1, job.job_2]
}

job "job_1" {
  name = "job 1"
  steps = [step.step_1]
}

job "job_2" {
  name = "job 2"

  runs {
    on = "ubuntu-20.04"
  }

  needs = [job.job_1]

  steps = [step.step_2]
}

step "step_1" {
  name = "step 1"
}

step "step_2" {
  name = "step 2"
}
`)

	var expect_1 = []byte(`on:
  label:
    types:
    - created
  push:
    branches:
    - main
    tags:
    - v2
jobs:
  job_1:
    name: job 1
    steps:
      step_1:
        id: step_1
        name: step 1
  job_2:
    name: job 2
    needs:
    - job_1
    runs-on: ubuntu-20.04
    steps:
      step_2:
        id: step_2
        name: step 2
`)

	var have_2 = []byte(`workflow "workflow_1" {
  filename = "dummy-file.yaml"
  on {
    event "push" {
      branches = ["main"]
      tags = ["v2"]
    }
  }
  on {
    activity "label" {
      types = ["created"]
    } 
  }
  jobs = [job.job_1, job.job_2]
}

job "job_1" {
  name = "job 1"
  steps = [step.step_1]
}

job "job_2" {
  name = "job 2"

  runs {
    on = "ubuntu-20.04"
  }

  needs = [job.job_1]

  steps = [step.step_2]
}

step "step_1" {
  name = "step 1"
}

step "step_2" {
  name = "step 2"
}`)

	var expect_2 = []byte(`on:
  label:
    types:
    - created
  push:
    branches:
    - main
    tags:
    - v2
jobs:
  job_1:
    name: job 1
    steps:
      step_1:
        id: step_1
        name: step 1
  job_2:
    name: job 2
    needs:
    - job_1
    runs-on: ubuntu-20.04
    steps:
      step_2:
        id: step_2
        name: step 2
`)

	var tests = []ParserTest{
		{"workflow event push with a job", have_1, expect_1},
		{"workflow for atos itself", have_2, expect_2},
	}

	flags := atosFlag.NewParseFlags()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filename := "dummy-file"

			hclFile := fmt.Sprintf("%s/%s.hcl", tempDir, filename)
			yamlFile := fmt.Sprintf("%s/%s.yaml", tempDir, filename)

			atosWriter := writer.New()
			if err := atosWriter.Do(hclFile, have_1); err != nil {
				t.Fatal(err.Error())
			}

			flags.File = hclFile

			if err := do(flags, tempDir); err != nil {
				t.Fatal(err.Error())
			}

			file, err := os.ReadFile(yamlFile)
			if err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(file, tt.expect) {
				t.Fatalf("expected %s but got %s", tt.expect, file)
			}
		})
	}
}
