// TODO: this is to be removed after moving all tests to the new approach

package job

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/yldio/atos/internal/reader"
)

type TestingTimeoutMinutes struct {
	TimeoutMinutes uint16 `yaml:"timeout-minutes"`
}

type TestingStrategy struct {
	Strategy any `yaml:"strategy"`
}

type TestingContinueOnError struct {
	ContinueOnError any `yaml:"continue-on-error"`
}

type TestingContainer struct {
	Container any `yaml:"container"`
}

type TestingUses struct {
	Uses any `yaml:"uses"`
}

// TODO: remove in favor of parsers.HelperConvertHcl for now
func HelperConvertHcl(src []byte, val any) error {
	atosReader := reader.NewReader("dummy-directory", "dummy-file.hcl", false)
	body, err := atosReader.ReadHclSrc(src, "dummy-file.hcl")
	if err != nil {
		return err
	}

	diags := gohcl.DecodeBody(body, nil, val)
	if diags.HasErrors() {
		return err
	}

	return nil
}
