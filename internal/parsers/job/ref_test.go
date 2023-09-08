package job

import (
	"reflect"
	"testing"
)

func TestParseRef(t *testing.T) {
	t.Run("convert from hcl: ref", func(t *testing.T) {
		have := []byte(`job "job_1" {}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				Ref string `hcl:",label"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := hclConfig.Jobs[0].Ref

		expected := "job_1"

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: name", func(t *testing.T) {
		have := TestingRef{
			Ref: "job_1",
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`{}
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
