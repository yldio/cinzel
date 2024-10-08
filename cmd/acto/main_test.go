// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package main

import (
	"testing"
)

func TestParseHcl(t *testing.T) {
	// 	type ParserTest struct {
	// 		name   string
	// 		have   []byte
	// 		expect []byte
	// 	}

	// 	var have1 = []byte(`variable "name" {
	//   value = "workflow name"
	// }

	// workflow "workflow1" {
	//   filename = "dummy-file"

	//   name = variable.name

	//   on {
	//     event "push" {
	//       branches = ["main"]
	//       tags = ["v2"]
	//     }
	//   }

	//   on {
	//     activity "label" {
	//       types = ["created"]
	//     }
	//   }

	//   jobs = [job.job1, job.job2]
	// }

	// job "job1" {
	//   name = "job 1"
	//   steps = [step.step1]

	//   runs {
	//     on = "ubuntu-20.04"
	//   }
	// }

	// job "job2" {
	//   name = "job 2"

	//   runs {
	//     on = "ubuntu-20.04"
	//   }

	//   needs = [job.job1]

	//   steps = [step.step2]
	// }

	// step "step1" {
	//   name = "step 1"
	//   run = "echo \"step 1\""
	// }

	// step "step2" {
	//   name = "step 2"
	//   run = "echo \"step 2\""
	// }
	// `)

	// 	var expect1 = []byte(`name: workflow name
	// on:
	//   label:
	//     types:
	//     - created
	//   push:
	//     branches:
	//     - main
	//     tags:
	//     - v2
	// jobs:
	//   job1:
	//     name: job 1
	//     runs-on: ubuntu-20.04
	//     steps:
	//     - id: step1
	//       name: step 1
	//       run: echo "step 1"
	//   job2:
	//     name: job 2
	//     needs:
	//     - job1
	//     runs-on: ubuntu-20.04
	//     steps:
	//     - id: step2
	//       name: step 2
	//       run: echo "step 2"
	// `)

	// 	var have2 = []byte(`workflow "workflow2" {
	//   filename = "dummy-file"

	//   on {
	//     event "push" {
	//       branches = ["main"]
	//       tags = ["v2"]
	//     }
	//   }

	//   on {
	//     activity "label" {
	//       types = ["created"]
	//     }
	//   }

	//   jobs = [job.job1, job.job2]
	// }

	// job "job1" {
	//   name = "job 1"
	//   steps = [step.step1]

	//   runs {
	//     on = "ubuntu-20.04"
	//   }
	// }

	// job "job2" {
	//   name = "job 2"

	//   runs {
	//     on = "ubuntu-20.04"
	//   }

	//   needs = [job.job1]

	//   steps = [step.step2]
	// }

	// step "step1" {
	//   name = "step 1"
	//   run = "echo \"step 1\""
	// }

	// step "step2" {
	//   name = "step 2"
	//   run = "echo \"step 2\""
	// }`)

	// 	var expect2 = []byte(`on:
	//   label:
	//     types:
	//     - created
	//   push:
	//     branches:
	//     - main
	//     tags:
	//     - v2
	// jobs:
	//   job1:
	//     name: job 1
	//     runs-on: ubuntu-20.04
	//     steps:
	//     - id: step1
	//       name: step 1
	//       run: echo "step 1"
	//   job2:
	//     name: job 2
	//     needs:
	//     - job1
	//     runs-on: ubuntu-20.04
	//     steps:
	//     - id: step2
	//       name: step 2
	//       run: echo "step 2"
	// `)

	// 	var tests = []ParserTest{
	// 		{"workflow event push with a job", have1, expect1},
	// 		{"workflow for acto itself", have2, expect2},
	// 	}

	// 	flags := actoflag.New()

	// 	for _, tt := range tests {
	// 		t.Run(tt.name, func(t *testing.T) {
	// 			tempDir := t.TempDir()
	// 			filename := "dummy-file"

	// 			hclFile := fmt.Sprintf("%s/%s.hcl", tempDir, filename)
	// 			yamlFile := fmt.Sprintf("%s/%s.yaml", tempDir, filename)

	// 			actoWriter := filewriter.New()
	// 			if err := actoWriter.Do(hclFile, tt.have); err != nil {
	// 				t.Fatal(err.Error())
	// 			}

	// 			flags.File = hclFile

	// 			if err := do(flags, tempDir); err != nil {
	// 				t.Fatal(err.Error())
	// 			}

	// 			file, err := os.ReadFile(yamlFile)
	// 			if err != nil {
	// 				t.Fatal(err.Error())
	// 			}

	//			if !reflect.DeepEqual(file, tt.expect) {
	//				t.Fatalf("expected \n%s but got \n%s", tt.expect, file)
	//			}
	//		})
	//	}
}
