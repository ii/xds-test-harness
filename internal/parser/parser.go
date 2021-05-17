package parser

import (
	"gopkg.in/yaml.v2"
)

func ParseYaml(yml string) (*EnvoyConfig, error) {
	var config EnvoyConfig

	err := yaml.Unmarshal([]byte(yml), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
