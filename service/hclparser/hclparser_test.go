// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package hclparser

import (
	"errors"
	"reflect"
	"testing"

	"github.com/yldio/acto/internal/action"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/job"
	"github.com/yldio/acto/internal/step"
	"github.com/yldio/acto/internal/workflow"
	"github.com/yldio/acto/service/reader"
	"github.com/yldio/acto/service/yamlparser"
)

func TestParseHcl(t *testing.T) {
	type ParserTest struct {
		name   string
		have   string
		expect *yamlparser.Yaml
	}

	var have1 = `workflow "workflow1" {
  filename = "dummy-file"
  on {
    events = "push"
  }
  jobs = [job.job1]
}

job "job1" {
  name = "job 1"

  runs {
    on = "ubuntu-latest"
  }

  steps = [step.step1]
}

step "step1" {
  run = "echo \"step1\""
}
`
	var job1 = "job 1"
	var runsOn any = "ubuntu-latest"
	var run1 = "echo \"step1\""
	var expect1 = yamlparser.New(workflow.Workflows{
		{
			Id:       "workflow1",
			Filename: "dummy-file",
			On:       action.On("push"),
			JobsIds:  []string{"job1"},
			Jobs: map[string]job.Job{
				"job1": {
					Id:     "job1",
					Name:   &job1,
					RunsOn: &runsOn,
					Steps: step.Steps{
						{
							Id:  "step1",
							Run: &run1,
						},
					},
					StepsIds: []string{"step1"},
				},
			},
		},
	})

	var have2 = `workflow "workflow2" {
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
  jobs = [job.job2]
}

job "job2" {
  name = "job 2"

  runs {
    on = "ubuntu-latest"
  }

  steps = [step.step2]
}

step "step2" {
  run = "echo \"step2\""
}
`
	var job2 = "job 2"
	var run2 = "echo \"step2\""
	var expect2 = yamlparser.New(workflow.Workflows{
		{
			Id:       "workflow2",
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
			JobsIds: []string{"job2"},
			Jobs: map[string]job.Job{
				"job2": {
					Id:     "job2",
					Name:   &job2,
					RunsOn: &runsOn,
					Steps: step.Steps{
						{
							Id:  "step2",
							Run: &run2,
						},
					},
					StepsIds: []string{"step2"},
				},
			},
		},
	})

	var have3 = `workflow "workflow3" {
  on {
    events = "push"
  }
  jobs = [job.job3]
}

job "job3" {
  name = "job 3"

  runs {
    on = "ubuntu-latest"
  }

  steps = [step.step3]
}

step "step3" {
  run = "echo \"step3\""
}
`

	var tests = []ParserTest{
		{"workflow event push with a job", have1, expect1},
		{"workflow events push with filters and activity label and with a job", have2, expect2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// easier to use actoReader to read HCL than to create an hcl.Body variable to pass to NewHclParse
			actoReader := reader.New("dummy-file.hcl", false)

			hclBody, err := actoReader.ReadHclSrc([]byte(tt.have), "dummy-file.hcl")
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
		// easier to use actoReader to read HCL than to create an hcl.Body variable to pass to NewHclParse
		actoReader := reader.New("dummy-file.hcl", false)

		hclBody, err := actoReader.ReadHclSrc([]byte(have3), "dummy-file.hcl")
		if err != nil {
			t.Fatal(err.Error())
		}

		parser := New(hclBody)

		if err := parser.Decode(); err != nil {
			t.Fatal(err.Error())
		}

		_, err = parser.Parse()
		if !errors.Is(err, actoerrors.ErrWorkflowFilenameRequired) {
			t.Fatalf("error message should be %s", actoerrors.ErrWorkflowFilenameRequired)
		}
	})
}
