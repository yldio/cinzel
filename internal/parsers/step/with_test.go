package step

import (
	"reflect"
	"testing"
)

func TestWith(t *testing.T) {
	t.Run("convert from hcl: with", func(t *testing.T) {
		have := []byte(`step {
  with {
    name  = "first_name"
    value = "Mona"
  }

  with {
    name  = "middle_name"
    value = "The"
  }

  with {
    name  = "last_name"
    value = "Octocat"
  }
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				With WithsConfig `hcl:"with,block"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := hclConfig.Steps[0].With

		expected := WithsConfig{
			{
				Name:  "first_name",
				Value: "Mona",
			},
			{
				Name:  "middle_name",
				Value: "The",
			},
			{
				Name:  "last_name",
				Value: "Octocat",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: with", func(t *testing.T) {
		have := WithsConfig{
			{
				Name:  "first_name",
				Value: "Mona",
			},
			{
				Name:  "middle_name",
				Value: "The",
			},
			{
				Name:  "last_name",
				Value: "Octocat",
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := map[string]any{
			"first_name":  "Mona",
			"middle_name": "The",
			"last_name":   "Octocat",
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: permissions", func(t *testing.T) {
		have := TestingWith{
			With: map[string]any{
				"first_name":  "Mona",
				"middle_name": "The",
				"last_name":   "Octocat",
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`with:
  first_name: Mona
  last_name: Octocat
  middle_name: The
`,
		)
		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
