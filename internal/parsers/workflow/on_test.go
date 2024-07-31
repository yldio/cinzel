package workflow

import (
	"reflect"
	"testing"
)

func TestParseOn(t *testing.T) {
	t.Run("convert from hcl: event push", func(t *testing.T) {
		have := []byte(`workflow {
  on = "push"
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				On *OnConfig `hcl:"on,attr"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Workflows[0].On

		expected := OnConfig("push")

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: event push", func(t *testing.T) {
		have := OnConfig("push")

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := "push"

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: event push", func(t *testing.T) {
		have := TestingOn{
			On: "push",
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`on: push
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
