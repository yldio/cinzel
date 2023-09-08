package job

import (
	"reflect"
	"testing"
)

func TestParseTimeoutMinutes(t *testing.T) {
	t.Run("convert from hcl: timeout-minutes", func(t *testing.T) {
		have := []byte(`job {
  timeout_minutes = 5
}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				TimeoutMinutes *TimeoutMinutesConfig `hcl:"timeout_minutes,attr"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Jobs[0].TimeoutMinutes

		expected := TimeoutMinutesConfig(5)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: timeout-minutes", func(t *testing.T) {
		have := TimeoutMinutesConfig(5)

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := uint16(5)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: timeout-minutes", func(t *testing.T) {
		have := TestingTimeoutMinutes{
			TimeoutMinutes: uint16(5),
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`timeout-minutes: 5
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
