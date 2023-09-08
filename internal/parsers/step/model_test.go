package step

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/yldio/atos/internal/reader"
)

type TestingRef struct {
	Ref string `yaml:"-"`
}

type TestingId struct {
	Id string `yaml:"id"`
}

type TestingIf struct {
	If string `yaml:"if"`
}

type TestingName struct {
	Name string `yaml:"name"`
}

type TestingUses struct {
	Uses string `yaml:"uses"`
}

type TestingRun struct {
	Run string `yaml:"run"`
}

type TestingWorkingDirectory struct {
	WorkingDirectory string `yaml:"working-directory"`
}

type TestingShell struct {
	Shell string `yaml:"shell"`
}

type TestingWith struct {
	With map[string]any `yaml:"with"`
}

type TestingEnv struct {
	Env map[string]any `yaml:"env"`
}

type TestingContinueOnError struct {
	ContinueOnError bool `yaml:"continue-on-error"`
}

type TestingTimeoutMinutes struct {
	TimeoutMinutes uint16 `yaml:"timeout-minutes"`
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
