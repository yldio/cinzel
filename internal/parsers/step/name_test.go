package step

import (
	"reflect"
	"testing"
)

func TestParseName(t *testing.T) {
	t.Run("convert from hcl: name", func(t *testing.T) {
		have := []byte(`step {
  name = "Print a greeting"
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Name *NameConfig `hcl:"name,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Steps[0].Name

		expected := NameConfig("Print a greeting")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: name", func(t *testing.T) {
		have := NameConfig("Print a greeting")

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := "Print a greeting"

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: name", func(t *testing.T) {
		have := TestingName{
			Name: "Print a greeting",
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`name: Print a greeting
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
