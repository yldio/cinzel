package workflow

import (
	"reflect"
	"testing"
)

func TestParseRunName(t *testing.T) {
	t.Run("convert from hcl: run-name", func(t *testing.T) {
		// TODO: document that HCL has special escape sequences such as the `$${` for `${`.
		have := []byte(`workflow {
  run_name = "Deploy to $${{ inputs.deploy_target }} by @$${{ github.actor }}"
}
`,
		)

		var hclConfig struct {
			Workflows []struct {
				RunName *RunNameConfig `hcl:"run_name,attr"`
			} `hcl:"workflow,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.FailNow()
		}

		got := *hclConfig.Workflows[0].RunName

		expected := RunNameConfig("Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}")

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("parse from hcl: run-name", func(t *testing.T) {
		have := RunNameConfig("Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}")

		got, err := have.Parse()
		if err != nil {
			t.FailNow()
		}

		expected := "Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}"

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})

	t.Run("convert to yaml: name", func(t *testing.T) {
		have := TestingRunName{
			RunName: "Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}",
		}

		got, err := Convert(have)
		if err != nil {
			t.FailNow()
		}

		expected := []byte(`run-name: Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.FailNow()
		}
	})
}
