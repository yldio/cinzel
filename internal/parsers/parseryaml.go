package parsers

import (
	"bytes"

	"github.com/goccy/go-yaml"
	"github.com/yldio/atos/internal/parsers/actions"
)

type YamlParser struct {
	yaml    []actions.WorkflowYaml
	rawYaml [][]byte
}

func NewYamlParser() *YamlParser {
	return &YamlParser{}
}

func (parse *YamlParser) ParseToYaml(workflows []actions.Workflow) error {
	for _, workflow := range workflows {
		yaml, err := workflow.ConvertToYaml()
		if err != nil {
			return err
		}

		parse.yaml = append(parse.yaml, yaml)
	}

	return nil
}

func (parse *YamlParser) Do() error {
	var rawYamls [][]byte
	for _, y := range parse.yaml {
		content, err := yaml.Marshal(y)
		if err != nil {
			return err
		}

		rawYamls = append(rawYamls, content)
	}

	parse.rawYaml = rawYamls

	return nil
}

func (parse *YamlParser) GetContent() [][]byte {
	return parse.rawYaml
}

func Convert(content any) ([]byte, error) {
	out, err := yaml.Marshal(content)
	if err != nil {
		return []byte{}, err
	}

	// Please link to https://github.com/go-yaml/yaml?tab=readme-ov-file#yaml-support-for-the-go-language
	// `atos` uses `any` so we need this "hack" to clean `"on":` to just `on:`.
	filteredOut := bytes.Replace(out, []byte("\"on\""), []byte("on"), -1)

	return filteredOut, nil
}
