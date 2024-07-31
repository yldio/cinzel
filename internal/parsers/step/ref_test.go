package step

import (
	"reflect"
	"testing"
)

func TestParseRef(t *testing.T) {
	t.Run("convert from hcl: ref", func(t *testing.T) {
		have := []byte(`step "step_1" {}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Ref string `hcl:",label"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := hclConfig.Steps[0].Ref

		expected := "step_1"

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: name", func(t *testing.T) {
		have := TestingRef{
			Ref: "step_1",
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`{}
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
