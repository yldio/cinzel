package workflow

import (
	"reflect"
	"testing"
)

func TestParsePushBranches(t *testing.T) {
	t.Run("convert from hcl: event push with activity tags", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "push"
    filter = "tags"
    values = ["v2", "v1.*"]
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

		event := "push"
		filter := "tags"
		values := []string{"v2", "v1.*"}

		expected := OnByFilterConfig{
			Event:  event,
			Filter: &filter,
			Values: &values,
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert from hcl: event push with activity branches and tags-ignore", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "push"
    filter = "branches"
    values = ["main", "mona/octocat", "releases/**"]
  }
  
  on_by_filter {
    event  = "push"
    filter = "tags-ignore"
    values = ["v2", "v1.*"]
  }
  
  on_by_filter {
    event  = "push"
    filter = "paths"
    values = ["**.js"]
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

		got := hclConfig.Workflows[0].OnByFilter

		event := "push"
		filter_1 := "branches"
		values_1 := []string{"main", "mona/octocat", "releases/**"}
		filter_2 := "tags-ignore"
		values_2 := []string{"v2", "v1.*"}
		filter_3 := "paths"
		values_3 := []string{"**.js"}

		expected := []*OnByFilterConfig{
			{
				Event:  event,
				Filter: &filter_1,
				Values: &values_1,
			},
			{
				Event:  event,
				Filter: &filter_2,
				Values: &values_2,
			},
			{
				Event:  event,
				Filter: &filter_3,
				Values: &values_3,
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: event pull_request with activity branches", func(t *testing.T) {
		event := "pull_request"
		filter := "branches"
		values := []string{"main", "mona/octocat", "releases/**"}

		have := OnByFilterConfig{
			Event:  event,
			Filter: &filter,
			Values: &values,
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := map[string]any{
			"pull_request": map[string][]string{
				"branches": {"main", "mona/octocat", "releases/**"},
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: event pull_request with activity branches and paths", func(t *testing.T) {
		have := TestingOn{
			On: map[string]any{
				"pull_request": map[string][]string{
					"branches": {"main", "mona/octocat", "releases/**"},
					"paths":    {"**.js"},
				},
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`on:
  pull_request:
    branches:
    - main
    - mona/octocat
    - releases/**
    paths:
    - "**.js"
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
