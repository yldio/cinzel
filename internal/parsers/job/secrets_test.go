package job

import (
	"reflect"
	"testing"
)

func TestParseSecrets(t *testing.T) {

	t.Run("convert from hcl: secrets", func(t *testing.T) {
		have := []byte(`job {
  secret {
    name = "access-token"
    value = "$${{ secrets.PERSONAL_ACCESS_TOKEN }}"
  }

  secret {
    name = "password"
    value = "$${{ secrets.PASSWORD }}"
  }
}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				Secret SecretsConfig `hcl:"secret,block"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := hclConfig.Jobs[0].Secret

		expected := SecretsConfig{
			{
				Name:  "access-token",
				Value: "${{ secrets.PERSONAL_ACCESS_TOKEN }}",
			},
			{
				Name:  "password",
				Value: "${{ secrets.PASSWORD }}",
			},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert from hcl: secrets inherit", func(t *testing.T) {
		have := []byte(`job {
  secrets = "inherit"
}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				Secrets SecretsInheritConfig `hcl:"secrets,attr"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := hclConfig.Jobs[0].Secrets

		expected := SecretsInheritConfig("inherit")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: secrets", func(t *testing.T) {
		have := SecretsConfig{
			{
				Name:  "access-token",
				Value: "${{ secrets.PERSONAL_ACCESS_TOKEN }}",
			},
			{
				Name:  "password",
				Value: "${{ secrets.PASSWORD }}",
			},
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := Secrets{
			"access-token": "${{ secrets.PERSONAL_ACCESS_TOKEN }}",
			"password":     "${{ secrets.PASSWORD }}",
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: secrets inherit", func(t *testing.T) {
		have := SecretsInheritConfig("inherit")

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := SecretsInherit("inherit")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: secrets", func(t *testing.T) {
		have := TestingSecrets{
			Secrets{
				"access-token": "${{ secrets.PERSONAL_ACCESS_TOKEN }}",
				"password":     "${{ secrets.PASSWORD }}",
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`secrets:
  access-token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
  password: ${{ secrets.PASSWORD }}
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: secrets inherit", func(t *testing.T) {
		have := TestingSecrets{
			SecretsInherit("inherit"),
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`secrets: inherit
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
