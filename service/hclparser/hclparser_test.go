// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package hclparser

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/action"
	"github.com/yldio/atos/internal/job"
	"github.com/yldio/atos/internal/workflow"
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
  filename = "dummy-file.yaml"
  on {
    events = "push"
  }
  jobs = [job.job_1]
}

job "job_1" {
  name = "job 1"
}
`
	var job_1 = "job 1"
	var expect_1 = yamlparser.New(workflow.Workflows{
		{
			Id:       "workflow_1",
			Filename: "dummy-file.yaml",
			On:       action.On("push"),
			JobsIds:  []string{"job_1"},
			Jobs: map[string]job.Job{
				"job_1": {
					Id:   "job_1",
					Name: &job_1,
				},
			},
		},
	})

	var have_2 = `workflow "workflow_2" {
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
  jobs = [job.job_2]
}

job "job_2" {
  name = "job 2"
}
`
	var job_2 = "job 2"
	var expect_2 = yamlparser.New(workflow.Workflows{
		{
			Id:       "workflow_2",
			Filename: "dummy-file.yaml",
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
					Id:   "job_2",
					Name: &job_2,
				},
			},
		},
	})

	var tests = []ParserTest{
		{"workflow event push with a job", have_1, expect_1},
		{"workflow events push with filters and activity label and with a job", have_2, expect_2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// easier to use atosReader to read HCL than to create an hcl.Body variable to pass to NewHclParse
			atosReader := reader.New("dummy-directory", "dummy-file.hcl", false)

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
}
