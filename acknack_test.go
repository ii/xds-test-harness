package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/cucumber/godog"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	pb "github.com/zachmandeville/tester-prototype/api/adapter"
	"github.com/zachmandeville/tester-prototype/internal/parser"
)

const configFile string = "config.yaml"

var (
	opts []grpc.DialOption = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}
)

type ClientConfig struct {
	Port   string
	Conn   *grpc.ClientConn
	NodeId string
}

type Runner struct {
	Adapter           *ClientConfig
	Target            *ClientConfig
	DiscoveryResponse *envoy_service_discovery_v3.DiscoveryResponse
	CDS               struct {
		Stream         cluster_service.ClusterDiscoveryService_StreamClustersClient
		Responses      *envoy_service_discovery_v3.DiscoveryResponse
		DeltaStream    cluster_service.ClusterDiscoveryService_DeltaClustersClient
		DeltaResponses *envoy_service_discovery_v3.DeltaDiscoveryResponse
	}
}

func (r *Runner) addPorts(*godog.Scenario) {
	configYaml, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading setup file: %v", err)
	}
	config, err := parser.ParseInitConfig(configYaml)
	if err != nil {
		fmt.Printf("Error parsing yaml file: %v", err)
	}
	r.Adapter = &ClientConfig{
		Port: config.Adapter,
	}
	r.Target = &ClientConfig{
		Port: config.Target,
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

func (r *Runner) aCDSStreamWasInitiatedWithADiscoveryRequestMatchingYaml(arg1 *godog.DocString) error {
	drdata, err := parser.ParseDiscoveryRequest(arg1.Content)
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
	stream, err := c.StreamClusters(context.Background())
	if err != nil {
		err = fmt.Errorf("Error starting CDS stream: %v\n", err)
		return err
	}
	r.CDS.Stream = stream
	r.CDS.Stream.Send(dreq)
	res, err := stream.Recv()
	r.CDS.Responses = res
	return nil
}

func (r *Runner) iSubscribeToDeltaCDSForTheseClusters(clustersYaml *godog.DocString) error {
	clusters, err := parser.ParseClusters(clustersYaml.Content)
	if err != nil {
		err = fmt.Errorf("Error parsing clusters: %v", err)
		return err
	}
	deltaReq := &envoy_service_discovery_v3.DeltaDiscoveryRequest{
		Node:                    &envoy_config_core_v3.Node{Id: "test-id"},
		ResourceNamesSubscribe:  clusters,
		InitialResourceVersions: map[string]string{},
		ResponseNonce:           "",
	}
	c := cluster_service.NewClusterDiscoveryServiceClient(r.Target.Conn)
	stream, err := c.DeltaClusters(context.Background())
	if err != nil {
		err = fmt.Errorf("Error starting stream: %v\n", err)
		return err
	}
	r.CDS.DeltaStream = stream
	r.CDS.DeltaStream.Send(deltaReq)
	res, err := stream.Recv()
	if err != nil {
		err = fmt.Errorf("error receiving clusters: %v\n", err)
		return err
	}
	r.CDS.DeltaResponses = res
	return nil
}

func (r *Runner) theStreamWasACKedWithADiscoveryRequestMatchingYaml(arg1 *godog.DocString) error {
	drdata, err := parser.ParseDiscoveryRequest(arg1.Content)
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
	r.CDS.Stream.Send(dreq)
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

func (r *Runner) iGetADiscoveryResponseMatchingYaml(arg1 *godog.DocString) error {
	var expected parser.DiscoveryResponse
	if err := yaml.Unmarshal([]byte(arg1.Content), &expected); err != nil {
		return err
	}
	actual, _ := parser.ParseDiscoveryResponse(r.DiscoveryResponse)
	if !reflect.DeepEqual(expected, *actual) {
		return fmt.Errorf("expected yaml does not match actual, %v vs. %v", expected, *actual)
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

func (r *Runner) isReachableViaGRPC(server string) error {
	switch server {
	case "adapter":
		err := connectViaGRPC(r.Adapter, server)
		return err
	case "target":
		err := connectViaGRPC(r.Target, server)
		return err
	default:
		err := fmt.Errorf("unexpected server name: %v", server)
		return err
	}
}

func (r *Runner) nodeidOfIs(server, nodeID string) error {
	switch server {
	case "target":
		r.Target.NodeId = nodeID
	case "adapter":
		r.Adapter.NodeId = nodeID
	default:
		err := fmt.Errorf("unexecpected server name: %v", server)
		return err
	}
	return nil
}

func (r *Runner) targetIsUpdatedToMatchYaml(yml *godog.DocString) error {
	snapshot, err := parser.YamlToSnapshot(yml.Content)
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

func (r *Runner) theClientReceivesADiscoveryResponseMatchingYaml(yml *godog.DocString) error {
	dreq := &envoy_service_discovery_v3.DiscoveryRequest{
		VersionInfo: "1",
		Node: &envoy_config_core_v3.Node{
			Id: "test-id",
		},
		ResourceNames: []string{},
		TypeUrl:       "type.googleapis.com/envoy.config.cluster.v3.Cluster",
		ResponseNonce: "1",
	}
	r.CDS.Stream.Send(dreq)
	res, err := r.CDS.Stream.Recv()
	if err != nil {
		err = fmt.Errorf("error receiving discovery response: %v\n", err)
		return err
	}
	var expected parser.DiscoveryResponse
	if err := yaml.Unmarshal([]byte(yml.Content), &expected); err != nil {
		return err
	}
	actual, _ := parser.ParseDiscoveryResponse(res)
	if !reflect.DeepEqual(expected, *actual) {
		return fmt.Errorf("expected yaml does not match actual, %v vs. %v", expected, *actual)
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	runner := &Runner{}

	ctx.BeforeScenario(runner.addPorts)
	ctx.Step(`^node-id of "([^"]*)" is "([^"]*)"$`, runner.nodeidOfIs)
	ctx.Step(`^a Target setup with snapshot matching yaml:$`, runner.aTargetSetupWithSnapshotMatchingYaml)
	ctx.Step(`^I subscribe to delta CDS for these clusters:$`, runner.iSubscribeToDeltaCDSForTheseClusters)
	ctx.Step(`^I get a discovery response matching yaml:$`, runner.iGetADiscoveryResponseMatchingYaml)
	ctx.Step(`^I send a discovery request matching yaml:$`, runner.iSendADiscoveryRequestMatchingYaml)
	ctx.Step(`^"([^"]*)" is reachable via gRPC$`, runner.isReachableViaGRPC)
	ctx.Step(`^Target is updated to match yaml:$`, runner.targetIsUpdatedToMatchYaml)
	ctx.Step(`^the client receives a discovery response matching yaml:$`, runner.theClientReceivesADiscoveryResponseMatchingYaml)
	ctx.Step(`^a CDS stream was initiated with a discovery request matching yaml:$`, runner.aCDSStreamWasInitiatedWithADiscoveryRequestMatchingYaml)
	ctx.Step(`^the stream was ACKed with a discovery request matching yaml:$`, runner.theStreamWasACKedWithADiscoveryRequestMatchingYaml)
}
