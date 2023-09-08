package workflow

import (
	"reflect"
	"testing"
)

func TestParseOnSchedule(t *testing.T) {
	t.Run("convert from hcl: event schedule with multiple crons", func(t *testing.T) {
		have := []byte(`workflow {
  on_by_filter {
    event  = "schedule"
    values = ["30 5 * * 1,3", "30 5 * * 2,4"]
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

		event := "schedule"
		values := []string{"30 5 * * 1,3", "30 5 * * 2,4"}

		expected := OnByFilterConfig{
			Event:  event,
			Values: &values,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: event schedule with multiple crons", func(t *testing.T) {
		event := "schedule"
		values := []string{"30 5 * * 1,3", "30 5 * * 2,4"}

		have := OnByFilterConfig{
			Event:  event,
			Values: &values,
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := map[string]any{
			"schedule": []map[string]string{
				{"cron": "30 5 * * 1,3"},
				{"cron": "30 5 * * 2,4"},
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: event schedule with multiple crons", func(t *testing.T) {
		have := TestingOn{
			On: map[string]any{
				"schedule": []map[string]string{
					{"cron": "30 5 * * 1,3"},
					{"cron": "30 5 * * 2,4"},
				},
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`on:
  schedule:
  - cron: 30 5 * * 1,3
  - cron: 30 5 * * 2,4
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
