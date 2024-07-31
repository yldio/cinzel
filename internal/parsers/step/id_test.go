package step

import (
	"reflect"
	"testing"
)

func TestParseId(t *testing.T) {
	t.Run("convert from hcl: id", func(t *testing.T) {
		have := []byte(`step {
  id = "step_1"
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Id *IdConfig `hcl:"id,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Steps[0].Id

		expected := IdConfig("step_1")

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: id", func(t *testing.T) {
		have := IdConfig("step_1")

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := "step_1"

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: id", func(t *testing.T) {
		have := TestingId{
			Id: "step_1",
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`id: step_1
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
