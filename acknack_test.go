package main

import (
	// "context"
	// "encoding/json"
	"fmt"
	"io/ioutil"
	// "log"
	// "reflect"
	"time"
	"github.com/cucumber/godog"
	// envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	// cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	// "github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"
	// pb "github.com/zachmandeville/tester-prototype/api/adapter"
	"github.com/zachmandeville/tester-prototype/internal/parser"
)

const configFile string = "config.yaml"

var (
	opts []grpc.DialOption
)

type ClientConfig struct {
	Port string
	Conn *grpc.ClientConn
}

type Runner struct {
	Adapter           *ClientConfig
	Target            *ClientConfig
	TargetAdapter     *ClientConfig
	DiscoveryResponse *envoy_service_discovery_v3.DiscoveryResponse
}

func (r *Runner) addPorts(*godog.Scenario)  {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading setup file: %v", err)
	}
	yamlConfig, err := parser.ParseInitConfig(yamlFile)
	if err != nil {
		fmt.Printf("Error parsing yaml file: %v", err)
	}
	r.Adapter = &ClientConfig{}
	r.Adapter.Port = yamlConfig.Adapter
	r.Target = &ClientConfig{}
	r.Target.Port = yamlConfig.Target
	r.TargetAdapter = &ClientConfig{}
	r.TargetAdapter.Port = yamlConfig.TargetAdapter
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

func connectViaGRPC (client *ClientConfig, server string) error {
		conn, err := grpc.Dial(client.Port, opts...)
		if err != nil {
			err = fmt.Errorf("Cannot connect to %v: %v", server, err)
			return err
		}
		client.Conn = conn
		return nil
}

func (r *Runner) isReachableViaGrpc(server string) error {
	switch server {
	case "adapter":
		err := connectViaGRPC(r.Adapter, server)
		return err
	case "target":
		err := connectViaGRPC(r.Target, server)
		return err
	case "target_adapter":
		err := connectViaGRPC(r.TargetAdapter, server)
		return err
	default:
		return godog.ErrPending
	}
}


func InitializeScenario(ctx *godog.ScenarioContext) {
	runner := &Runner{}
	ctx.BeforeScenario(runner.addPorts)
	ctx.BeforeScenario(func(s *godog.Scenario) {
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithBlock())
		opts = append(opts, grpc.WithTimeout(time.Second*5))
	})
	ctx.Step(`^a Target setup with snapshot matching yaml:$`, aTargetSetupWithSnapshotMatchingYaml)
	ctx.Step(`^I get a discovery response matching json:$`, iGetADiscoveryResponseMatchingJson)
	ctx.Step(`^I send a discovery request matching yaml:$`, iSendADiscoveryRequestMatchingYaml)
	ctx.Step(`^"([^"]*)" is reachable via grpc$`, runner.isReachableViaGrpc)
}
