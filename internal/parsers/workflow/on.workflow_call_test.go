package workflow

import (
	"reflect"
	"testing"
)

func TestParseWorkflowCall(t *testing.T) {
	t.Run("convert from hcl: workflow_call with input, output and secret", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "workflow_call"

    input "username" {
      type        = "string"
      description = "A username passed from the caller workflow"
      default     = "john-doe"
      required    = false
    }

    output "workflow_output1" {
      description = "The first job output"
      value       = "$${{ jobs.my_job.outputs.job_output1 }}"
    }

    secret "access-token" {
      description = "A token passed from the caller workflow"
      required    = false
    }
  }
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				OnByFilter []*OnByFilterConfig `hcl:"on_by_filter,block"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Workflows[0].OnByFilter[0]

		event := "workflow_call"

		descriptionInput := "A username passed from the caller workflow"
		defaultInput := "john-doe"
		requiredInput := false

		input := OnInput{
			Name:        "username",
			Type:        "string",
			Description: &descriptionInput,
			Default:     &defaultInput,
			Required:    &requiredInput,
		}

		descriptionOutput := "The first job output"
		valueOutput := "${{ jobs.my_job.outputs.job_output1 }}"

		output := OnOutput{
			Name:        "workflow_output1",
			Description: &descriptionOutput,
			Value:       &valueOutput,
		}

		descriptionSecret := "A token passed from the caller workflow"
		requiredSecret := false

		secret := OnSecret{
			Name:        "access-token",
			Description: &descriptionSecret,
			Required:    &requiredSecret,
		}

		expected := OnByFilterConfig{
			Event: event,
			Input: []*OnInput{
				&input,
			},
			Output: []*OnOutput{
				&output,
			},
			Secret: []*OnSecret{
				&secret,
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: workflow_call with input, output and secret", func(t *testing.T) {
		event := "workflow_call"

		descriptionInput := "A username passed from the caller workflow"
		defaultInput := "john-doe"
		requiredInput := false

		input := OnInput{
			Name:        "username",
			Type:        "string",
			Description: &descriptionInput,
			Default:     &defaultInput,
			Required:    &requiredInput,
		}

		descriptionOutput := "The first job output"
		valueOutput := "${{ jobs.my_job.outputs.job_output1 }}"

		output := OnOutput{
			Name:        "workflow_output1",
			Description: &descriptionOutput,
			Value:       &valueOutput,
		}

		descriptionSecret := "A token passed from the caller workflow"
		requiredSecret := false

		secret := OnSecret{
			Name:        "access-token",
			Description: &descriptionSecret,
			Required:    &requiredSecret,
		}

		have := OnByFilterConfig{
			Event: event,
			Input: []*OnInput{
				&input,
			},
			Output: []*OnOutput{
				&output,
			},
			Secret: []*OnSecret{
				&secret,
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := map[string]any{
			"workflow_call": map[string]any{
				"inputs": map[string]any{
					"username": &OnInput{
						Name:        "username",
						Type:        "string",
						Description: &descriptionInput,
						Default:     &defaultInput,
						Required:    &requiredInput,
					},
				},
				"outputs": map[string]any{
					"workflow_output1": &OnOutput{
						Name:        "workflow_output1",
						Description: &descriptionOutput,
						Value:       &valueOutput,
					},
				},
				"secrets": map[string]any{
					"access-token": &OnSecret{
						Name:        "access-token",
						Description: &descriptionSecret,
						Required:    &requiredSecret,
					},
				},
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: workflow_call with input, output and secret", func(t *testing.T) {
		descriptionInput := "A username passed from the caller workflow"
		defaultInput := "john-doe"
		requiredInput := false

		descriptionOutput := "The first job output"
		valueOutput := "${{ jobs.my_job.outputs.job_output1 }}"

		descriptionSecret := "A token passed from the caller workflow"
		requiredSecret := false

		have := TestingOn{
			On: map[string]any{
				"workflow_call": map[string]any{
					"inputs": map[string]any{
						"username": &OnInput{
							Name:        "username",
							Type:        "string",
							Description: &descriptionInput,
							Default:     &defaultInput,
							Required:    &requiredInput,
						},
					},
					"outputs": map[string]any{
						"workflow_output1": &OnOutput{
							Name:        "workflow_output1",
							Description: &descriptionOutput,
							Value:       &valueOutput,
						},
					},
					"secrets": map[string]any{
						"access-token": &OnSecret{
							Name:        "access-token",
							Description: &descriptionSecret,
							Required:    &requiredSecret,
						},
					},
				},
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`on:
  workflow_call:
    inputs:
      username:
        type: string
        description: A username passed from the caller workflow
        default: john-doe
        required: false
    outputs:
      workflow_output1:
        description: The first job output
        value: ${{ jobs.my_job.outputs.job_output1 }}
    secrets:
      access-token:
        description: A token passed from the caller workflow
        required: false
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
