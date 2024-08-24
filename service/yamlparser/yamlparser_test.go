// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package yamlparser

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/action"
	"github.com/yldio/atos/internal/job"
	"github.com/yldio/atos/internal/step"
	"github.com/yldio/atos/internal/workflow"
)

func TestParseYaml(t *testing.T) {
	type ParserTest struct {
		name   string
		have   *Yaml
		expect map[string][]byte
	}

	var job_1 = "job 1"
	var step_1 = "step 1"
	var have_1 = New(workflow.Workflows{
		{
			Id:       "workflow_1",
			Filename: "dummy-file",
			On:       action.On("push"),
			Jobs: map[string]job.Job{
				"job_1": {
					Id:       "job_1",
					Name:     &job_1,
					Needs:    &[]string{"job_2"},
					StepsIds: []string{"step_1"},
					Steps: step.Steps{
						{
							Id:   "step_1",
							Name: &step_1,
						},
					},
				},
				"job_2": {
					Id:       "job_2",
					StepsIds: []string{"step_1"},
					Steps: step.Steps{
						{
							Id:   "step_1",
							Name: &step_1,
						},
					},
				},
			},
		},
	})
	var expect_1 = map[string][]byte{
		"dummy-file.yaml": []byte(`on: push
jobs:
  job_1:
    name: job 1
    needs:
    - job_2
    steps:
    - id: step_1
      name: step 1
  job_2:
    steps:
    - id: step_1
      name: step 1
`),
	}

	var job_2 = "job 2"
	var have_2 = New(workflow.Workflows{
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
					Id:       "job_2",
					Name:     &job_2,
					StepsIds: []string{"step_1"},
					Steps: step.Steps{
						{
							Id:   "step_1",
							Name: &step_1,
						},
					},
				},
			},
		},
	})
	var expect_2 = map[string][]byte{
		"dummy-file.yaml": []byte(`on:
  label:
    types:
    - created
  push:
    branches:
    - main
    tags:
    - v2
jobs:
  job_2:
    name: job 2
    steps:
    - id: step_1
      name: step 1
`),
	}

	var tests = []ParserTest{
		{"a workflow with on push", have_1, expect_1},
		{"a workflow with on push with branches and tags", have_2, expect_2},
	}

	for _, tt := range tests {

		got, err := tt.have.Do()
		if err != nil {
			t.Fatal(err.Error())
		}

		if !reflect.DeepEqual(got, tt.expect) {
			t.Fatalf("expected %s but got %s", tt.expect, got)
		}
	}
}
