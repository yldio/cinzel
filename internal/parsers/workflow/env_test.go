package workflow

import (
	"reflect"
	"testing"
)

func TestEnv(t *testing.T) {
	t.Run("convert from hcl: env", func(t *testing.T) {
		have := []byte(`workflow {
  env {
    name  = "SERVER"
    value = "production"
  }

  env {
    name  = "TOKEN"
    value = "12345-abcde"
  }
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				Env EnvsConfig `hcl:"env,block"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := hclConfig.Workflows[0].Env

		expected := EnvsConfig{
			{
				Name:  "SERVER",
				Value: "production",
			},
			{
				Name:  "TOKEN",
				Value: "12345-abcde",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: env", func(t *testing.T) {
		have := EnvsConfig{
			{
				Name:  "SERVER",
				Value: "production",
			},
			{
				Name:  "TOKEN",
				Value: "12345-abcde",
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := map[string]any{
			"SERVER": "production",
			"TOKEN":  "12345-abcde",
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: env", func(t *testing.T) {
		have := TestingEnv{
			Env: map[string]any{
				"SERVER": "production",
				"TOKEN":  "12345-abcde",
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`env:
  SERVER: production
  TOKEN: 12345-abcde
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
