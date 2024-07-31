package job

import (
	"reflect"
	"testing"
)

func TestParseUses(t *testing.T) {

	t.Run("convert from hcl: uses", func(t *testing.T) {
		have := []byte(`job {
  uses = "octo-org/another-repo/.github/workflows/workflow.yml@v1"
}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				Uses UsesConfig `hcl:"uses,attr"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := hclConfig.Jobs[0].Uses

		expected := UsesConfig("octo-org/another-repo/.github/workflows/workflow.yml@v1")

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: uses", func(t *testing.T) {
		have := UsesConfig("./.github/workflows/workflow-2.yml")

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := Uses("./.github/workflows/workflow-2.yml")

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: uses", func(t *testing.T) {
		have := TestingUses{
			Uses("./.github/workflows/workflow-2.yml"),
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`uses: ./.github/workflows/workflow-2.yml
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
