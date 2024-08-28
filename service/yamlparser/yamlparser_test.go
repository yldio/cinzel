// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package yamlparser

import (
	"reflect"
	"testing"

	"github.com/yldio/acto/internal/action"
	"github.com/yldio/acto/internal/job"
	"github.com/yldio/acto/internal/step"
	"github.com/yldio/acto/internal/workflow"
)

func TestParseYaml(t *testing.T) {
	type ParserTest struct {
		name   string
		have   *Yaml
		expect map[string][]byte
	}

	var job1 = "job 1"
	var step1 = "step 1"
	var have1 = New(workflow.Workflows{
		{
			Id:       "workflow1",
			Filename: "dummy-file",
			On:       action.On("push"),
			Jobs: map[string]job.Job{
				"job1": {
					Id:       "job1",
					Name:     &job1,
					Needs:    &[]string{"job2"},
					StepsIds: []string{"step1"},
					Steps: step.Steps{
						{
							Id:   "step1",
							Name: &step1,
						},
					},
				},
				"job2": {
					Id:       "job2",
					StepsIds: []string{"step1"},
					Steps: step.Steps{
						{
							Id:   "step1",
							Name: &step1,
						},
					},
				},
			},
		},
	})
	var expect1 = map[string][]byte{
		"dummy-file.yaml": []byte(`on: push
jobs:
  job1:
    name: job 1
    needs:
    - job2
    steps:
    - id: step1
      name: step 1
  job2:
    steps:
    - id: step1
      name: step 1
`),
	}

	var job2 = "job 2"
	var have2 = New(workflow.Workflows{
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
					Id:       "job2",
					Name:     &job2,
					StepsIds: []string{"step1"},
					Steps: step.Steps{
						{
							Id:   "step1",
							Name: &step1,
						},
					},
				},
			},
		},
	})
	var expect2 = map[string][]byte{
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
  job2:
    name: job 2
    steps:
    - id: step1
      name: step 1
`),
	}

	var tests = []ParserTest{
		{"a workflow with on push", have1, expect1},
		{"a workflow with on push with branches and tags", have2, expect2},
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
