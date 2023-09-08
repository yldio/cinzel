package workflow

import (
	"reflect"
	"testing"
)

func TestParseOnAsList(t *testing.T) {
	t.Run("convert from hcl: event push and pull_request", func(t *testing.T) {
		have := []byte(`workflow {
  on_as_list = ["push", "pull_request"]
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				OnAsList *OnAsListConfig `hcl:"on_as_list,attr"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Workflows[0].OnAsList

		expected := OnAsListConfig{"push", "pull_request"}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: event push and pull_request", func(t *testing.T) {
		have := OnAsListConfig{"push", "pull_request"}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := []string{"push", "pull_request"}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: event push and pull_request", func(t *testing.T) {
		have := TestingOn{
			On: []string{"push", "pull_request"},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`on:
- push
- pull_request
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
