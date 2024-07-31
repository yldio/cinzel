package step

import (
	"reflect"
	"testing"
)

func TestParseUses(t *testing.T) {
	t.Run("convert from hcl: uses", func(t *testing.T) {
		have := []byte(`step {
  uses {
    action  = "actions/checkout"
	version = "v4"
  }
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Uses *UsesConfig `hcl:"uses,block"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Steps[0].Uses

		expected := UsesConfig{
			Action:  "actions/checkout",
			Version: "v4",
		}

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: uses", func(t *testing.T) {
		have := UsesConfig{
			Action:  "actions/checkout",
			Version: "v4",
		}

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := "actions/checkout@v4"

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: uses", func(t *testing.T) {
		have := TestingUses{
			Uses: "actions/checkout@v4",
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`uses: actions/checkout@v4
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
