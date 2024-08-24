// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package hclparser

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/action"
	"github.com/yldio/atos/internal/job"
	"github.com/yldio/atos/internal/step"
	"github.com/yldio/atos/internal/workflow"
	"github.com/yldio/atos/service/atoserrors"
	"github.com/yldio/atos/service/reader"
	"github.com/yldio/atos/service/yamlparser"
)

func TestParseHcl(t *testing.T) {
	type ParserTest struct {
		name   string
		have   string
		expect *yamlparser.Yaml
	}

	var have_1 = `workflow "workflow_1" {
  filename = "dummy-file"
  on {
    events = "push"
  }
  jobs = [job.job_1]
}

job "job_1" {
  name = "job 1"

  runs {
    on = "ubuntu-latest"
  }

  steps = [step.step_1]
}

step "step_1" {
  run = "echo \"step_1\""
}
`
	var job_1 = "job 1"
	var runsOn any = "ubuntu-latest"
	var run_1 = "echo \"step_1\""
	var expect_1 = yamlparser.New(workflow.Workflows{
		{
			Id:       "workflow_1",
			Filename: "dummy-file",
			On:       action.On("push"),
			JobsIds:  []string{"job_1"},
			Jobs: map[string]job.Job{
				"job_1": {
					Id:     "job_1",
					Name:   &job_1,
					RunsOn: &runsOn,
					Steps: step.Steps{
						{
							Id:  "step_1",
							Run: &run_1,
						},
					},
					StepsIds: []string{"step_1"},
				},
			},
		},
	})

	var have_2 = `workflow "workflow_2" {
  filename = "dummy-file"
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
  jobs = [job.job_2]
}

job "job_2" {
  name = "job 2"

  runs {
    on = "ubuntu-latest"
  }

  steps = [step.step_2]
}

step "step_2" {
  run = "echo \"step_2\""
}
`
	var job_2 = "job 2"
	var run_2 = "echo \"step_2\""
	var expect_2 = yamlparser.New(workflow.Workflows{
		{
			Id:       "workflow_2",
			Filename: "dummy-file",
			On: action.On(map[string]map[string][]string{
				"push": {
					"branches": []string{"main"},
					"tags":     []string{"v2"},
				},
				"label": {
					"types": []string{"created"},
				},
			}),
			JobsIds: []string{"job_2"},
			Jobs: map[string]job.Job{
				"job_2": {
					Id:     "job_2",
					Name:   &job_2,
					RunsOn: &runsOn,
					Steps: step.Steps{
						{
							Id:  "step_2",
							Run: &run_2,
						},
					},
					StepsIds: []string{"step_2"},
				},
			},
		},
	})

	var have_3 = `workflow "workflow_3" {
  on {
    events = "push"
  }
  jobs = [job.job_3]
}

job "job_3" {
  name = "job 3"

  runs {
    on = "ubuntu-latest"
  }

  steps = [step.step_3]
}

step "step_3" {
  run = "echo \"step_3\""
}
`

	var tests = []ParserTest{
		{"workflow event push with a job", have_1, expect_1},
		{"workflow events push with filters and activity label and with a job", have_2, expect_2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// easier to use atosReader to read HCL than to create an hcl.Body variable to pass to NewHclParse
			atosReader := reader.New("dummy-file.hcl", false)

			hclBody, err := atosReader.ReadHclSrc([]byte(tt.have), "dummy-file.hcl")
			if err != nil {
				t.Fatal(err.Error())
			}

			parser := New(hclBody)

			if err := parser.Decode(); err != nil {
				t.Fatal(err.Error())
			}

			content, err := parser.Parse()
			if err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(content, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}

	t.Run("workflow fails because no Filename", func(t *testing.T) {
		// easier to use atosReader to read HCL than to create an hcl.Body variable to pass to NewHclParse
		atosReader := reader.New("dummy-file.hcl", false)

		hclBody, err := atosReader.ReadHclSrc([]byte(have_3), "dummy-file.hcl")
		if err != nil {
			t.Fatal(err.Error())
		}

		parser := New(hclBody)

		if err := parser.Decode(); err != nil {
			t.Fatal(err.Error())
		}

		_, err = parser.Parse()
		if !errors.Is(err, atoserrors.ErrWorkflowFilenameRequired) {
			t.Fatalf("error message should be %s", atoserrors.ErrWorkflowFilenameRequired)
		}
	})
}
