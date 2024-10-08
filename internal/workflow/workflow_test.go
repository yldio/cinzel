// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package workflow

// import (
// 	"reflect"
// 	"testing"

// 	"github.com/hashicorp/hcl/v2"
// 	"github.com/hashicorp/hcl/v2/hclsyntax"
// 	"github.com/yldio/acto/internal/action"
// )

// func TestWorkflows(t *testing.T) {
// 	type Test struct {
// 		name   string
// 		have   *WorkflowsConfig
// 		expect Workflows
// 	}

// 	filename1, _ := hclsyntax.ParseExpression([]byte(`"dummy-file"`), "", hcl.Pos{})
// 	event1, _ := hclsyntax.ParseExpression([]byte(`"push"`), "", hcl.Pos{})
// 	jobs1, _ := hclsyntax.ParseExpression([]byte(`[job.job1]`), "", hcl.Pos{})

// 	var have1 = WorkflowsConfig{
// 		{
// 			Id:       "workflow1",
// 			Filename: filename1,
// 			On: action.OnsConfig{
// 				{
// 					Events: event1,
// 				},
// 			},
// 			Jobs: jobs1,
// 		},
// 	}
// 	var expect1 = Workflows{
// 		{
// 			Id:       "workflow1",
// 			Filename: "dummy-file",
// 			On:       "push",
// 			JobsIds: []string{
// 				"job1",
// 			},
// 		},
// 	}

// 	var tests = []Test{
// 		{"with defined Id", &have1, expect1},
// 		{"with defined Id, on and jobs", &have1, expect1},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.have.Parse()
// 			if err != nil {
// 				t.Fatal(err.Error())
// 			}

// 			if !reflect.DeepEqual(got, tt.expect) {
// 				t.Fatal(tt.name)
// 			}
// 		})
// 	}
// }
