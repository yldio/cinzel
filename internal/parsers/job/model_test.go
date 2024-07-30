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

type TestingServices struct {
	Services any `yaml:"services"`
}

type TestingUses struct {
	Uses any `yaml:"uses"`
}

type TestingWith struct {
	With any `yaml:"with"`
}

type TestingSecrets struct {
	Secrets any `yaml:"secrets"`
}

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
