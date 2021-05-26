package main

import (
	// "context"
	// "encoding/json"
	"fmt"
	"io/ioutil"
	// "log"
	// "reflect"
	// "time"
	"github.com/cucumber/godog"
	"gopkg.in/yaml.v2"
	// envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	// cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	// "github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"
	// pb "github.com/zachmandeville/tester-prototype/api/adapter"
	// "github.com/zachmandeville/tester-prototype/internal/parser"
)

type yamlconfig struct {
	address       string `yaml:"address"`
	target        string `yaml:"target"`
	targetadapter string `yaml:"target_address"`
}

type ClientConfig struct {
	Port string
	Conn *grpc.ClientConn
}

type Runner struct {
	Adapter           *ClientConfig
	Target            *ClientConfig
	Shim              *ClientConfig
	DiscoveryResponse *envoy_service_discovery_v3.DiscoveryResponse
}

func aTargetSetupWithSnapshotMatchingYaml(arg1 *godog.DocString) error {
	return godog.ErrPending
}

func iGetADiscoveryResponseMatchingJson(arg1 *godog.DocString) error {
	return godog.ErrPending
}

func iSendADiscoveryRequestMatchingYaml(arg1 *godog.DocString) error {
	return godog.ErrPending
}

func isReachableViaGrpc(server string) error {
	switch server {
	case "adapter":
		return godog.ErrPending
	case "target":
		return godog.ErrPending
	case "target_adapter":
		return godog.ErrPending
	default:
		return godog.ErrPending
	}
}

func init() {
	yamlFile, err := ioutil.ReadFile("setup.yaml")
	if err != nil {
		fmt.Printf("Error reading setup file", err)
	}
	var yamlConfig YamlConfig
	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		fmt.Printf("Error parsing yaml file", err)
	}
	fmt.Println(yamlConfig)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a Target setup with snapshot matching yaml:$`, aTargetSetupWithSnapshotMatchingYaml)
	ctx.Step(`^I get a discovery response matching json:$`, iGetADiscoveryResponseMatchingJson)
	ctx.Step(`^I send a discovery request matching yaml:$`, iSendADiscoveryRequestMatchingYaml)
	ctx.Step(`^"([^"]*)" is reachable via grpc$`, runner.isReachableViaGrpc)
}
