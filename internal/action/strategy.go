// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

type MatrixVariable struct {
	Name  string
	Value any
}

type MatrixVariableConfig struct {
	Name  hcl.Expression `hcl:"name,attr"`
	Value hcl.Expression `hcl:"value,attr"`
}

type MatrixVariablesConfig []MatrixVariableConfig

type MatrixConfig struct {
	Variables MatrixVariablesConfig `hcl:"variable,block"`
	Include   hcl.Expression        `hcl:"include,attr"`
	Exclude   hcl.Expression        `hcl:"exclude,attr"`
}

type StrategyConfig struct {
	Matrix      MatrixConfig   `hcl:"matrix,block"`
	FailFast    hcl.Expression `hcl:"fail_fast,attr"`
	MaxParallel hcl.Expression `hcl:"max_parallel,attr"`
}

type Matrix map[string]any

type Strategy struct {
	Matrix      *Matrix `yaml:"matrix" hcl:"matrix"`
	FailFast    *bool   `yaml:"fail-fast,omitempty" hcl:"fail_fast"`
	MaxParallel *uint64 `yaml:"max-parallel,omitempty" hcl:"max_parallel"`
}

func (config *StrategyConfig) unwrapMaxParallel(acto *actoparser.Acto) (*uint64, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case int64:
		if resultValue < 0 {
			return nil, errors.New("attribute 'max_parallel' must be a positive number")
		}

		val := uint64(resultValue)
		return &val, nil
	case uint64:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapMaxParallel(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'max_parallel' must be a positive number")
	}
}

func (config *StrategyConfig) parseMaxParallel() (*uint64, error) {
	acto := actoparser.NewActo(config.MaxParallel)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapMaxParallel(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *StrategyConfig) unwrapFailFast(acto *actoparser.Acto) (*bool, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapFailFast(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'fail_fast' must be a boolean")
	}
}

func (config *StrategyConfig) parseFailFast() (*bool, error) {
	acto := actoparser.NewActo(config.FailFast)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapFailFast(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *MatrixConfig) unwrapMatrixExclude(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []map[string]any:
		return resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapMatrixExclude(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'exclude' must be a list objects")
	}
}

func (config *MatrixConfig) unwrapMatrixInclude(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []map[string]any:
		return resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapMatrixInclude(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'include' must be a list objects")
	}
}

func (config *MatrixVariableConfig) unwrapMatrixValue(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'value' must be a list of strings, numbers, boleans or objects")
	case []string:
		return resultValue, nil
	case []bool:
		return resultValue, nil
	case []int64:
		return resultValue, nil
	case []uint64:
		return resultValue, nil
	case []float64:
		return resultValue, nil
	case []map[string]any:
		return resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapMatrixValue(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'value' must be a list of strings, numbers, boleans or objects")
	}
}

func (config *MatrixVariableConfig) parseMatrixValue() (any, error) {
	acto := actoparser.NewActo(config.Value)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapMatrixValue(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *MatrixVariableConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'name' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapName(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'name' must be a string")
	}
}

func (config *MatrixVariableConfig) parseMatrixName() (*string, error) {
	acto := actoparser.NewActo(config.Name)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapName(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *MatrixConfig) parseMatrixInclude() (any, error) {
	acto := actoparser.NewActo(config.Include)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapMatrixInclude(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *MatrixConfig) parseMatrixExclude() (any, error) {
	acto := actoparser.NewActo(config.Exclude)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapMatrixExclude(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *MatrixVariableConfig) parseMatrixVariable() (*MatrixVariable, error) {
	name, err := config.parseMatrixName()
	if err != nil {
		return nil, err
	}

	value, err := config.parseMatrixValue()
	if err != nil {
		return nil, err
	}

	variable := MatrixVariable{
		Name:  *name,
		Value: value,
	}

	return &variable, nil
}

func (config *StrategyConfig) parseMatrix() (*Matrix, error) {
	matrix := make(Matrix)

	for _, matrixVariable := range config.Matrix.Variables {
		variable, err := matrixVariable.parseMatrixVariable()
		if err != nil {
			return nil, err
		}

		matrix[variable.Name] = variable.Value
	}

	include, err := config.Matrix.parseMatrixInclude()
	if err != nil {
	}

	if include != nil {
		matrix["include"] = include
	}

	exclude, err := config.Matrix.parseMatrixExclude()
	if err != nil {
	}

	if exclude != nil {
		matrix["exclude"] = exclude
	}

	return &matrix, nil
}

func (config *StrategyConfig) Parse() (*Strategy, error) {
	if config == nil {
		return nil, nil
	}

	strategy := Strategy{}

	matrix, err := config.parseMatrix()
	if err != nil {
		return nil, fmt.Errorf("error in strategy: %w, %w", err, actoerrors.ErrOpenIssue)
	}

	strategy.Matrix = matrix

	failFast, err := config.parseFailFast()
	if err != nil {
		return nil, fmt.Errorf("error in strategy: %w", err)
	}

	if failFast != nil {
		strategy.FailFast = failFast
	}

	maxParallel, err := config.parseMaxParallel()
	if err != nil {
		return nil, fmt.Errorf("error in strategy: %w", err)
	}

	if maxParallel != nil {
		strategy.MaxParallel = maxParallel
	}

	return &strategy, nil
}

func (strategy *Strategy) Decode(body *hclwrite.Body, attr string) error {
	if len(body.Blocks()) > 0 {
		body.AppendNewline()
	}

	strategyBlock := body.AppendNewBlock(attr, nil)
	strategyBody := strategyBlock.Body()

	if strategy.Matrix != nil {
		matrixAttr, err := actoparser.GetHclTag(*strategy, "Matrix")
		if err != nil {
			return err
		}

		matrixBlock := strategyBody.AppendNewBlock(matrixAttr, nil)
		matrixBody := matrixBlock.Body()

		for varK, varV := range *strategy.Matrix {
			if varK == "include" {
				switch v := varV.(type) {
				case []any:
					var list []cty.Value
					for _, vv := range v {
						switch vvv := vv.(type) {
						case string:
							list = append(list, cty.StringVal(vvv))
						case int64:
							list = append(list, cty.NumberIntVal(vvv))
						case uint64:
							list = append(list, cty.NumberUIntVal(vvv))
						case float64:
							list = append(list, cty.NumberFloatVal(vvv))
						case bool:
							list = append(list, cty.BoolVal(vvv))
						case map[string]any:
							subList := make(map[string]cty.Value)
							for a, b := range vvv {
								switch c := b.(type) {
								case string:
									subList[a] = cty.StringVal(c)
								case int64:
									subList[a] = cty.NumberIntVal(c)
								case uint64:
									subList[a] = cty.NumberUIntVal(c)
								case float64:
									subList[a] = cty.NumberFloatVal(c)
								case bool:
									subList[a] = cty.BoolVal(c)
								}
							}
							list = append(list, cty.ObjectVal(subList))
						}
					}
					matrixBody.SetAttributeValue("include", cty.TupleVal(list))
				default:
					panic("missing dealt type")
				}
			} else if varK == "exclude" {
				switch v := varV.(type) {
				case []any:
					var list []cty.Value
					for _, vv := range v {
						switch vvv := vv.(type) {
						case string:
							list = append(list, cty.StringVal(vvv))
						case int64:
							list = append(list, cty.NumberIntVal(vvv))
						case uint64:
							list = append(list, cty.NumberUIntVal(vvv))
						case float64:
							list = append(list, cty.NumberFloatVal(vvv))
						case bool:
							list = append(list, cty.BoolVal(vvv))
						case map[string]any:
							subList := make(map[string]cty.Value)
							for a, b := range vvv {
								switch c := b.(type) {
								case string:
									subList[a] = cty.StringVal(c)
								case int64:
									subList[a] = cty.NumberIntVal(c)
								case uint64:
									subList[a] = cty.NumberUIntVal(c)
								case float64:
									subList[a] = cty.NumberFloatVal(c)
								case bool:
									subList[a] = cty.BoolVal(c)
								default:
									return errors.New("missing dealt type")
								}
							}
							list = append(list, cty.ObjectVal(subList))
						}
					}
					matrixBody.SetAttributeValue("exclude", cty.TupleVal(list))
				default:
					return errors.New("missing dealt type")
				}
			} else {
				variableBlock := matrixBody.AppendNewBlock("variable", nil)
				variableBody := variableBlock.Body()

				switch v := varV.(type) {
				case []any:
					var list []cty.Value
					for _, vv := range v {
						switch vvv := vv.(type) {
						case string:
							list = append(list, cty.StringVal(vvv))
						case int64:
							list = append(list, cty.NumberIntVal(vvv))
						case uint64:
							list = append(list, cty.NumberUIntVal(vvv))
						case float64:
							list = append(list, cty.NumberFloatVal(vvv))
						case bool:
							list = append(list, cty.BoolVal(vvv))
						case map[string]any:
							subList := make(map[string]cty.Value)
							for a, b := range vvv {
								switch c := b.(type) {
								case string:
									subList[a] = cty.StringVal(c)
								case int64:
									subList[a] = cty.NumberIntVal(c)
								case uint64:
									subList[a] = cty.NumberUIntVal(c)
								case float64:
									subList[a] = cty.NumberFloatVal(c)
								case bool:
									subList[a] = cty.BoolVal(c)
								default:
									return errors.New("missing dealt type")
								}
							}
							list = append(list, cty.ObjectVal(subList))
						}
					}
					variableBody.SetAttributeValue("value", cty.StringVal(varK))
					variableBody.SetAttributeValue("name", cty.TupleVal(list))
				default:
					return errors.New("missing dealt type")
				}
			}
		}
	}

	if strategy.FailFast != nil {
		attr, err := actoparser.GetHclTag(*strategy, "FailFast")
		if err != nil {
			return err
		}

		if len(strategyBody.Blocks()) > 0 {
			strategyBody.AppendNewline()
		}

		strategyBody.SetAttributeValue(attr, cty.BoolVal(*strategy.FailFast))
	}

	if strategy.MaxParallel != nil {
		attr, err := actoparser.GetHclTag(*strategy, "MaxParallel")
		if err != nil {
			return err
		}

		if len(strategyBody.Blocks()) > 0 {
			strategyBody.AppendNewline()
		}

		strategyBody.SetAttributeValue(attr, cty.NumberUIntVal(*strategy.MaxParallel))
	}

	return nil
}
