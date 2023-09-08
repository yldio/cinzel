package workflow

import (
	"reflect"
	"testing"
)

func TestParseName(t *testing.T) {
	t.Run("convert from hcl: name", func(t *testing.T) {
		have := []byte(`workflow {
  name = "Deploy to development"
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				Name *NameConfig `hcl:"name,attr"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Workflows[0].Name

		expected := NameConfig("Deploy to development")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: name", func(t *testing.T) {
		have := NameConfig("Deploy to development")

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := "Deploy to development"

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: name", func(t *testing.T) {
		have := TestingName{
			Name: "Deploy to development",
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`name: Deploy to development
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
