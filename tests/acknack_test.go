package main

import (
	"context"
	"encoding/json"
	"fmt"
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

var (
	opts []grpc.DialOption
)

type runner struct {
	adapter *grpc.ClientConn
	target  *grpc.ClientConn
	shim    *grpc.ClientConn
	discoveryResponse *envoy_service_discovery_v3.DiscoveryResponse
}

func (r *runner) aTargetWithClustersSpecifiedWithYaml(yml *godog.DocString) error {
	var specifiedClusters []string
	spec, err := parser.ParseYaml(yml.Content)
	if err != nil {
		log.Printf("error parsing yaml file: %+v", err)
		return err
	}
	for _, c := range spec.Clusters {
		specifiedClusters = append(specifiedClusters, c.Name)
	}

	c := pb.NewAdapterClient(r.adapter)
	resource := &pb.ResourceSpec{
		Spec: yml.Content,
	}

	snapshot, err := c.RegisterResource(context.Background(), resource)
	if err != nil {
		log.Printf("error registering resource: %v", err)
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(snapshot.Snapshot), &data); err != nil {
		panic(err)
	}

	clusters := data["Resources"].([]interface{})[1].(map[string]interface{})["Items"]
	var snapshotClusters []string
	for key, _ := range clusters.(map[string]interface{}) {
		snapshotClusters = append(snapshotClusters, key)
	}
	if !reflect.DeepEqual(snapshotClusters, specifiedClusters) {
		return err
	}
	return nil
}

func (r *runner) aShimLocatedAt(port string) error {
	c := pb.NewAdapterClient(r.adapter)
	request := &pb.ShimRequest{
		Port: port,
	}
	_, err := c.ConnectToShim(context.Background(), request)
	if err != nil {
		log.Printf("error connecting to shim: %v", err)
		return err
	}
	return nil
}

func (r *runner) aTargetLocatedAt(port string) error {
	conn, err := grpc.Dial(port, opts...)
	if err != nil {
		log.Printf("error connecting to target: %v", err)
		return err
	}
	r.target = conn
	return nil
}

func (r *runner) anAdapterLocatedAt(port string) error {
	conn, err := grpc.Dial(port, opts...)
	if err != nil {
		log.Printf("error connecting to adapter: %v", err)
		return err
	}
	r.adapter = conn
	return nil
}

func (r *runner) iGetAResponseContaining(arg1 string) error {
	m := jsonpb.Marshaler{}
    result, _ := m.MarshalToString(r.discoveryResponse)
	log.Printf("\n\n\n%v\n\n\n\n", result)
	return godog.ErrPending
}

func (r *runner) iGetADiscoveryResponseMatchingJson(arg1 *godog.DocString) error {
	var expected, actual interface{}
	m := jsonpb.Marshaler{}
    result, _ := m.MarshalToString(r.discoveryResponse)
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


func (r *runner) iSendADiscoveryRequestMatchingYaml(arg1 *godog.DocString) error {
	drdata, err := parser.ParseDiscoveryRequest(arg1.Content)
	if err != nil{
	  log.Printf("error parsing discovery request: %v\n", err)
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
	c := cluster_service.NewClusterDiscoveryServiceClient(r.target)
	// TODO this should be a general fetch, not just to clusters.
	dres, err := c.FetchClusters(context.Background(),dreq)
	if err != nil {
		log.Printf("err fetching clusters: %v", err.Error())
		return err
	}
	r.discoveryResponse = dres
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	runner := &runner{}
	ctx.BeforeScenario(func(s *godog.Scenario) {
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithBlock())
		opts = append(opts, grpc.WithTimeout(time.Second*5))
	})
	ctx.Step(`^a Target with clusters specified with yaml:$`, runner.aTargetWithClustersSpecifiedWithYaml)
	ctx.Step(`^a Shim located at "([^"]*)"$`, runner.aShimLocatedAt)
	ctx.Step(`^a Target located at "([^"]*)"$`, runner.aTargetLocatedAt)
	ctx.Step(`^an Adapter located at "([^"]*)"$`, runner.anAdapterLocatedAt)
	ctx.Step(`^I send a discovery request matching yaml:$`, runner.iSendADiscoveryRequestMatchingYaml)
    ctx.Step(`^I get a discovery response matching json:$`, runner.iGetADiscoveryResponseMatchingJson)
}
