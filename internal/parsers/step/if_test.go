package step

import (
	"reflect"
	"testing"
)

func TestParseIf(t *testing.T) {
	t.Run("convert from hcl: if", func(t *testing.T) {
		have := []byte(`step {
  if = "$${{ ! startsWith(github.ref, 'refs/tags/') }}"
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				If *IfConfig `hcl:"if,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Steps[0].If

		expected := IfConfig("${{ ! startsWith(github.ref, 'refs/tags/') }}")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: if", func(t *testing.T) {
		have := IfConfig("${{ ! startsWith(github.ref, 'refs/tags/') }}")

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := IfConfig("${{ ! startsWith(github.ref, 'refs/tags/') }}")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: if", func(t *testing.T) {
		have := TestingIf{
			If: "${{ ! startsWith(github.ref, 'refs/tags/') }}",
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`if: ${{ ! startsWith(github.ref, 'refs/tags/') }}
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
