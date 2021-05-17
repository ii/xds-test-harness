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

func NewDiscoveryRequest() TestDiscoveryRequest {
	dr := TestDiscoveryRequest{}
	dr.VersionInfo = ""
	dr.ResponseNonce = ""
	dr.TypeURL = ""
	dr.VersionInfo = ""
	dr.Node.ID = ""
	return dr
}

func ParseDiscoveryRequest(yml string) (*TestDiscoveryRequest, error) {
	request := NewDiscoveryRequest()
	err := yaml.Unmarshal([]byte(yml), &request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
