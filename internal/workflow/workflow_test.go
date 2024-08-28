// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package workflow

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/acto/internal/action"
	"github.com/zclconf/go-cty/cty"
)

func TestWorkflows(t *testing.T) {
	type Test struct {
		name   string
		have   *WorkflowsConfig
		expect Workflows
	}

	var event1 = cty.StringVal("push")
	var exprList = hclsyntax.TupleConsExpr{
		Exprs: []hclsyntax.Expression{
			&hclsyntax.ScopeTraversalExpr{
				Traversal: hcl.Traversal{
					hcl.TraverseRoot{
						Name: "job",
					},
					hcl.TraverseAttr{
						Name: "job1",
					},
				},
			},
		},
		SrcRange:  hcl.Range{},
		OpenRange: hcl.Range{},
	}

	var filename = "dummy-file"
	var have1 = WorkflowsConfig{
		{
			Id:       "workflow1",
			Filename: &filename,
			On: action.OnsConfig{
				{
					Events: &event1,
				},
			},
			Jobs: &exprList,
		},
	}
	var expect1 = Workflows{
		{
			Id:       "workflow1",
			Filename: filename,
			On:       action.On("push"),
			JobsIds: []string{
				"job1",
			},
		},
	}

	var tests = []Test{
		{"with defined Id", &have1, expect1},
		{"with defined Id, on and jobs", &have1, expect1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Error(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}
}
