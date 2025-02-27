// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

type Credentials struct {
	Username string `yaml:"username,omitempty" hcl:"username"`
	Password string `yaml:"password,omitempty" hcl:"password"`
}

type CredentialsConfig struct {
	Username hcl.Expression `hcl:"username,attr"`
	Password hcl.Expression `hcl:"password,attr"`
}

func (config *CredentialsConfig) unwrapPassword(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'password' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapPassword(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'password' must be a string")
	}
}

func (config *CredentialsConfig) parsePassword() (*string, error) {
	acto := actoparser.NewActo(config.Password)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapPassword(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *CredentialsConfig) unwrapUsername(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'username' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapUsername(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'username' must be a string")
	}
}

func (config *CredentialsConfig) parseUsername() (*string, error) {
	acto := actoparser.NewActo(config.Username)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapUsername(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *CredentialsConfig) Parse() (*Credentials, error) {
	if config == nil {
		return nil, nil
	}

	credentials := Credentials{}

	username, err := config.parseUsername()
	if err != nil {
		return nil, err
	}

	if username != nil {
		credentials.Username = *username
	}

	password, err := config.parsePassword()
	if err != nil {
		return nil, err
	}

	if password != nil {
		credentials.Password = *password
	}

	return &credentials, nil
}

func (credentials *Credentials) Decode(body *hclwrite.Body, attr string) error {
	if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
		body.AppendNewline()
	}

	credentialsBlock := body.AppendNewBlock(attr, nil)
	credentialsBody := credentialsBlock.Body()

	if credentials.Username != "" {
		usernameAttr, err := actoparser.GetHclTag(*credentials, "Username")
		if err != nil {
			return err
		}

		credentialsBody.SetAttributeValue(usernameAttr, cty.StringVal(credentials.Username))
	}

	if credentials.Password != "" {
		passwordAttr, err := actoparser.GetHclTag(*credentials, "Password")
		if err != nil {
			return err
		}

		credentialsBody.SetAttributeValue(passwordAttr, cty.StringVal(credentials.Password))
	}

	return nil
}
