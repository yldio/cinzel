package workflow

import (
	"reflect"
	"testing"
)

func TestParseWorkflowDispatch(t *testing.T) {
	t.Run("convert from hcl: workflow_dispatch with branches", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "workflow_dispatch"

    input "logLevel" {
      type        = "choice"
      description = "Log level"
      default     = "warning"
      required    = true
      options     = ["info", "warning", "debug"]
    }

    input "print_tags" {
      type        = "boolean"
      description = "True to print to STDOUT"
      required    = true
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

		event := "workflow_dispatch"

		descriptionInput_1 := "Log level"
		defaultInput_1 := "warning"
		requiredInput_1 := true
		optionsInput_1 := []string{"info", "warning", "debug"}

		input_1 := OnInput{
			Name:        "logLevel",
			Type:        "choice",
			Description: &descriptionInput_1,
			Default:     &defaultInput_1,
			Required:    &requiredInput_1,
			Options:     &optionsInput_1,
		}

		descriptionInput_2 := "True to print to STDOUT"
		requiredInput_2 := true

		input_2 := OnInput{
			Name:        "print_tags",
			Type:        "boolean",
			Description: &descriptionInput_2,
			Required:    &requiredInput_2,
		}

		expected := OnByFilterConfig{
			Event: event,
			Input: []*OnInput{
				&input_1,
				&input_2,
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: workflow_run with branches", func(t *testing.T) {
		event := "workflow_dispatch"

		descriptionInput_1 := "Log level"
		defaultInput_1 := "warning"
		requiredInput_1 := true
		optionsInput_1 := []string{"info", "warning", "debug"}

		input_1 := OnInput{
			Name:        "logLevel",
			Type:        "choice",
			Description: &descriptionInput_1,
			Default:     &defaultInput_1,
			Required:    &requiredInput_1,
			Options:     &optionsInput_1,
		}

		descriptionInput_2 := "True to print to STDOUT"
		requiredInput_2 := true

		input_2 := OnInput{
			Name:        "print_tags",
			Type:        "boolean",
			Description: &descriptionInput_2,
			Required:    &requiredInput_2,
		}

		have := OnByFilterConfig{
			Event: event,
			Input: []*OnInput{
				&input_1,
				&input_2,
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := map[string]any{
			"workflow_dispatch": map[string]any{
				"inputs": map[string]any{
					"logLevel": &OnInput{
						Name:        "logLevel",
						Type:        "choice",
						Description: &descriptionInput_1,
						Required:    &requiredInput_1,
						Default:     &defaultInput_1,
						Options:     &optionsInput_1,
					},
					"print_tags": &OnInput{
						Name:        "print_tags",
						Type:        "boolean",
						Description: &descriptionInput_2,
						Required:    &requiredInput_2,
					},
				},
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: workflow_run with branches", func(t *testing.T) {
		descriptionInput_1 := "Log level"
		defaultInput_1 := "warning"
		requiredInput_1 := true
		optionsInput_1 := []string{"info", "warning", "debug"}

		descriptionInput_2 := "True to print to STDOUT"
		requiredInput_2 := true

		have := TestingOn{
			On: map[string]any{
				"workflow_dispatch": map[string]any{
					"inputs": map[string]any{
						"logLevel": &OnInput{
							Name:        "logLevel",
							Type:        "choice",
							Description: &descriptionInput_1,
							Required:    &requiredInput_1,
							Default:     &defaultInput_1,
							Options:     &optionsInput_1,
						},
						"print_tags": &OnInput{
							Name:        "print_tags",
							Type:        "boolean",
							Description: &descriptionInput_2,
							Required:    &requiredInput_2,
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
  workflow_dispatch:
    inputs:
      logLevel:
        type: choice
        description: Log level
        default: warning
        required: true
        options:
        - info
        - warning
        - debug
      print_tags:
        type: boolean
        description: True to print to STDOUT
        required: true
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
