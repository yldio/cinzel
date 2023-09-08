package workflow

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/yldio/atos/internal/reader"
)

type TestingName struct {
	Name string `yaml:"name"`
}

type TestingRunName struct {
	RunName string `yaml:"run-name"`
}

type TestingOn struct {
	On any `yaml:"on"`
}

type TestingPermissions struct {
	Permissions any `yaml:"permissions"`
}

type TestingEnv struct {
	Env any `yaml:"env"`
}

type TestingDefaults struct {
	Defaults any `yaml:"defaults"`
}

type TestingConcurrency struct {
	Concurrency any `yaml:"concurrency"`
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
