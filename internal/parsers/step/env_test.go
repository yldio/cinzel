package step

import (
	"reflect"
	"testing"
)

func TestEnv(t *testing.T) {
	t.Run("convert from hcl: env", func(t *testing.T) {
		have := []byte(`step {
  env {
    name  = "GITHUB_TOKEN"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }

  env {
    name  = "FIRST_NAME"
    value = "Mona"
  }

  env {
    name  = "LAST_NAME"
    value = "Octocat"
  }
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Env EnvsConfig `hcl:"env,block"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := hclConfig.Steps[0].Env

		expected := EnvsConfig{
			{
				Name:  "GITHUB_TOKEN",
				Value: "${{ secrets.GITHUB_TOKEN }}",
			},
			{
				Name:  "FIRST_NAME",
				Value: "Mona",
			},
			{
				Name:  "LAST_NAME",
				Value: "Octocat",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: env", func(t *testing.T) {
		have := EnvsConfig{
			{
				Name:  "GITHUB_TOKEN",
				Value: "${{ secrets.GITHUB_TOKEN }}",
			},
			{
				Name:  "FIRST_NAME",
				Value: "Mona",
			},
			{
				Name:  "LAST_NAME",
				Value: "Octocat",
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := map[string]any{
			"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
			"FIRST_NAME":   "Mona",
			"LAST_NAME":    "Octocat",
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: env", func(t *testing.T) {
		have := TestingEnv{
			Env: map[string]any{
				"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
				"FIRST_NAME":   "Mona",
				"LAST_NAME":    "Octocat",
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`env:
  FIRST_NAME: Mona
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  LAST_NAME: Octocat
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
