package job

import (
	"reflect"
	"testing"
)

func TestParseWith(t *testing.T) {

	t.Run("convert from hcl: with", func(t *testing.T) {
		have := []byte(`job {
  with {
    input {
      name = "username"
      value = "mona"
    } 

	input {
      name = "password"
      value = "octocat"
    } 
  }
}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				With WithConfig `hcl:"with,block"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := hclConfig.Jobs[0].With

		expected := WithConfig{
			[]WithInputConfig{
				{
					Name:  "username",
					Value: "mona",
				},
				{
					Name:  "password",
					Value: "octocat",
				},
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: with", func(t *testing.T) {
		have := WithConfig{
			[]WithInputConfig{
				{
					Name:  "username",
					Value: "mona",
				},
				{
					Name:  "password",
					Value: "octocat",
				},
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := With{
			"username": "mona",
			"password": "octocat",
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: with", func(t *testing.T) {
		have := TestingWith{
			With{
				"username": "mona",
				"password": "octocat",
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`with:
  password: octocat
  username: mona
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
