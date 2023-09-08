package workflow

import (
	"reflect"
	"testing"
)

func TestParseWorkflowRun(t *testing.T) {
	t.Run("convert from hcl: workflow_run with branches", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "workflow_run"

    workflows = ["Build"]
    types = ["requested"]
    branches = ["releases/**"]
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
			t.Fail()
		}

		got := *hclConfig.Workflows[0].OnByFilter[0]

		event := "workflow_run"
		workflows := []string{"Build"}
		types := []string{"requested"}
		branches := []string{"releases/**"}

		expected := OnByFilterConfig{
			Event:     event,
			Workflows: &workflows,
			Types:     &types,
			Branches:  &branches,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: workflow_run with branches", func(t *testing.T) {
		event := "workflow_run"
		workflows := []string{"Build"}
		types := []string{"requested"}
		branches := []string{"releases/**"}

		have := OnByFilterConfig{
			Event:     event,
			Workflows: &workflows,
			Types:     &types,
			Branches:  &branches,
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := map[string]any{
			"workflow_run": map[string][]string{
				"workflows": {"Build"},
				"types":     {"requested"},
				"branches":  {"releases/**"},
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: workflow_run with branches", func(t *testing.T) {
		have := TestingOn{
			On: map[string]any{
				"workflow_run": map[string][]string{
					"workflows": {"Build"},
					"types":     {"requested"},
					"branches":  {"releases/**"},
				},
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`on:
  workflow_run:
    branches:
    - releases/**
    types:
    - requested
    workflows:
    - Build
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
