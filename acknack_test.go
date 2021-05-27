package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/cucumber/godog"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/api/adapter"
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
	DiscoveryResponse *envoy_service_discovery_v3.DiscoveryResponse
}

func (r *Runner) addPorts(*godog.Scenario) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading setup file: %v", err)
	}
	yamlConfig, err := parser.ParseInitConfig(yamlFile)
	if err != nil {
		fmt.Printf("Error parsing yaml file: %v", err)
	}
	r.Adapter = &ClientConfig{
		Port: yamlConfig.Adapter,
	}
	r.Target = &ClientConfig{
		Port: yamlConfig.Target,
	}
}

func (r *Runner) aTargetSetupWithSnapshotMatchingYaml(snapYaml *godog.DocString) error {
	snapshot, err := parser.YamlToSnapshot(snapYaml.Content)
	if err != nil {
		err = fmt.Errorf("Error parsing snapshot yaml: %v", err)
		return err
	}

	c := pb.NewAdapterClient(r.Adapter.Conn)
	_, err = c.SetState(context.Background(), snapshot)
	if err != nil {
		err = fmt.Errorf("Cannot Set Target with State: %v\n", err)
		return err
	}
	return nil
}

func (r *Runner) iSendADiscoveryRequestMatchingYaml(dryaml *godog.DocString) error {
	drdata, err := parser.ParseDiscoveryRequest(dryaml.Content)
	if err != nil {
		err = fmt.Errorf("error parsing discovery request: %v\n", err)
		return err
	}
	dreq := &envoy_service_discovery_v3.DiscoveryRequest{
		VersionInfo: drdata.VersionInfo,
		Node: &envoy_config_core_v3.Node{
			Id: drdata.Node.ID,
		},
		ResourceNames: drdata.ResourceNames,
		TypeUrl:       drdata.TypeURL,
		ResponseNonce: drdata.ResponseNonce,
	}
	c := cluster_service.NewClusterDiscoveryServiceClient(r.Target.Conn)
	dres, err := c.FetchClusters(context.Background(), dreq)
	if err != nil {
		log.Printf("err fetching clusters: %v", err.Error())
		return err
	}
	r.DiscoveryResponse = dres
	return nil
}

func (r *Runner) iGetADiscoveryResponseMatchingJson(arg1 *godog.DocString) error {
	var expected, actual interface{}
	m := jsonpb.Marshaler{}
	result, _ := m.MarshalToString(r.DiscoveryResponse)
	if err := json.Unmarshal([]byte(result), &actual); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(arg1.Content), &expected); err != nil {
		return err
	}
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("expected JSON does not match actual, %v vs. %v", expected, actual)
	}
	return nil
}

func connectViaGRPC(client *ClientConfig, server string) error {
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
	default:
		return godog.ErrPending
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	runner := &Runner{}
	ctx.BeforeScenario(func(s *godog.Scenario) {
		opts = append(opts, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
	})
	ctx.BeforeScenario(runner.addPorts)
	ctx.Step(`^a Target setup with snapshot matching yaml:$`, runner.aTargetSetupWithSnapshotMatchingYaml)
	ctx.Step(`^I get a discovery response matching json:$`, runner.iGetADiscoveryResponseMatchingJson)
	ctx.Step(`^I send a discovery request matching yaml:$`, runner.iSendADiscoveryRequestMatchingYaml)
	ctx.Step(`^"([^"]*)" is reachable via grpc$`, runner.isReachableViaGrpc)
}
