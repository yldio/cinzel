package parsers

import (
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
