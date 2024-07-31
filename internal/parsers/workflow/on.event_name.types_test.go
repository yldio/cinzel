package workflow

import (
	"reflect"
	"testing"
)

func TestParseOnTypes(t *testing.T) {
	t.Run("convert from hcl: event label with activity types", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "label"
    filter = "types"
    values = ["created", "edited"]
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

		event := "label"
		filter := "types"
		values := []string{"created", "edited"}

		expected := OnByFilterConfig{
			Event:  event,
			Filter: &filter,
			Values: &values,
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: event label with activity types", func(t *testing.T) {
		event := "label"
		filter := "types"
		values := []string{"created", "edited"}

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
			"label": map[string][]string{
				"types": {"created", "edited"},
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: event label with activity types", func(t *testing.T) {
		have := TestingOn{
			On: map[string]any{
				"label": map[string][]string{
					"types": {"created", "edited"},
				},
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`on:
  label:
    types:
    - created
    - edited
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
