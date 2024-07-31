package workflow

import (
	"reflect"
	"testing"
)

func TestParsePullRequestBranches(t *testing.T) {
	t.Run("convert from hcl: event pull_request with activity branches", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "pull_request"
    filter = "branches"
    values = ["main", "mona/octocat", "releases/**"]
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

		event := "pull_request"
		filter := "branches"
		values := []string{"main", "mona/octocat", "releases/**"}

		expected := OnByFilterConfig{
			Event:  event,
			Filter: &filter,
			Values: &values,
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

	t.Run("convert to yaml: event pull_request with activity branches", func(t *testing.T) {
		have := TestingOn{
			On: map[string]any{
				"pull_request": map[string][]string{
					"branches": {"main", "mona/octocat", "releases/**"},
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
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
