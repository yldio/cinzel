// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"testing"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/internal/yamlwriter"
)

func TestStepParseSuccess(t *testing.T) {
	type Test struct {
		name string
		have []byte
		yaml string
	}

	var tests = []Test{
		{
			"test 1",
			[]byte(`step "my-step" {}`),
			"my-step:\n  id: my-step\n",
		},
		{
			"test 2",
			[]byte(`step "my-step" {
  id = "my-id"
  if = "$${{ failure() }}"
  name = "my-name"

  uses {
    action = "my-action"
    version = "v1.1.1"
  }

  timeout_minutes = 360
  continue_on_error = true
  working_directory = "./temp"
  shell = "bash"

  with {
    name  = "first_name"
	value = "Mona"
  }

  with {
    name  = "middle_name"
	value = "The"
  }

  with {
    name  = "last_name"
	value = "Octocat"
  }

  env {
    name  = "GITHUB_TOKEN"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }

  env {
    name  = "FIRST_NAME"
    value = "Mona"
  }

  env {
    name  = "LAST_NAME"
    value = "Octocat"
  }

  run = <<EOF
npm ci
npm run build
EOF
}`),
			`my-step:
  continue-on-error: true
  env:
    FIRST_NAME: Mona
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    LAST_NAME: Octocat
  id: my-id
  if: ${{ failure() }}
  name: my-name
  run: |
    npm ci
    npm run build
  shell: bash
  timeout-minutes: 360
  uses: my-action@v1.1.1
  with:
    first_name: Mona
    last_name: Octocat
    middle_name: The
  working-directory: ./temp
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			type HclConfig struct {
				Variables hclparser.VariablesConfig `hcl:"variable,block"`
				Steps     StepListConfig            `hcl:"step,block"`
			}

			hclConfig := HclConfig{}

			parser := hclparse.NewParser()

			hclFile, diags := parser.ParseHCL(tt.have, "example.hcl")
			if diags.HasErrors() {
				t.FailNow()
			}

			if diags = gohcl.DecodeBody(hclFile.Body, nil, &hclConfig); diags.HasErrors() {
				t.FailNow()
			}

			hv := hclparser.NewHCLVars()

			steps, err := hclConfig.Steps.Parse(hv)
			if err != nil {
				t.FailNow()
			}

			out, err := yamlwriter.Marshal(steps)
			if err != nil {
				t.FailNow()
			}

			if string(out) != tt.yaml {
				t.FailNow()
			}

		})
	}
}
