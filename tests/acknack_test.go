package main

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"time"

	"github.com/cucumber/godog"
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
	eq := reflect.DeepEqual(snapshotClusters, specifiedClusters)
	if eq == false {
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

func iGetAResponseContaining(arg1 string) error {
	return godog.ErrPending
}

func iSendAWildcardRequestToTheCDS() error {
	return godog.ErrPending
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
	ctx.Step(`^I get a response containing "([^"]*)"\.$`, iGetAResponseContaining)
	ctx.Step(`^I send a wildcard request to the CDS$`, iSendAWildcardRequestToTheCDS)
}
