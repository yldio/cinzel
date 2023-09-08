package step

import (
	"reflect"
	"testing"
)

func TestParseRun(t *testing.T) {
	t.Run("convert from hcl: run single line", func(t *testing.T) {
		have := []byte(`step {
  run = "npm install"
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Run *RunConfig `hcl:"run,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Steps[0].Run

		expected := RunConfig("npm install")

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert from hcl: run multi line", func(t *testing.T) {
		have := []byte(`step {
  run = <<-EOF
npm ci
npm run build
EOF
}
`,
		)

		var hclConfig struct {
			Steps []struct {
				Run *RunConfig `hcl:"run,attr"`
			} `hcl:"step,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Steps[0].Run

		expected := RunConfig(`npm ci
npm run build
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: run single line", func(t *testing.T) {
		have := RunConfig("npm install")

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := "npm install"

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: run multipe line", func(t *testing.T) {
		have := RunConfig(`npm ci
npm run build
`,
		)

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := `npm ci
npm run build
`

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: run single line", func(t *testing.T) {
		have := TestingRun{
			Run: "npm install",
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`run: npm install
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: run multi line", func(t *testing.T) {
		have := TestingRun{
			Run: `npm ci
npm run build
`,
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`run: |
  npm ci
  npm run build
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
