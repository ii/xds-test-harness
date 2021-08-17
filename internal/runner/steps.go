package runner

import (
	"context"
	"errors"
	"fmt"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/cucumber/godog"
	"github.com/ii/xds-test-harness/internal/parser"
	pb "github.com/zachmandeville/tester-prototype/api/adapter"
)

func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a target setup with the following state:$`, r.ATargetSetupWithTheFollowingState)
	ctx.Step(`^the Runner receives the following "([^"]*)":$`, r.TheRunnerReceivesTheFollowing)
	ctx.Step(`^the Runner sends its first CDS wildcard request to "([^"]*)"$`, r.TheRunnerSendsItsFirstCDSWildcardRequestTo)
}

func (r *Runner) ATargetSetupWithTheFollowingState(state *godog.DocString) error {
	snapshot, err := parser.YamlToSnapshot(state.Content)
	if err != nil {
		err = errors.New("Could not parse given state to adapter snapshot")
		return err
	}
	c := pb.NewAdapterClient(r.Adapter.Conn)
	_, err = c.SetState(context.Background(), snapshot)
	if err != nil {
		err = fmt.Errorf("Cannot Set Target with State: %v\n", err)
		return err
	}
	return err
}


func (r *Runner) TheRunnerSendsItsFirstCDSWildcardRequestTo(nodeID string) error {
	wildcardRequest := &discovery.DiscoveryRequest{
		VersionInfo: "1",
		Node: &envoy_config_core_v3.Node{
			Id: nodeID,
		},
		ResourceNames: []string{"*"},
		TypeUrl:       "type.googleapis.com/envoy.config.cluster.v3.Cluster",
	}
	responses := make(chan *discovery.DiscoveryResponse)
	errors := make(chan error)
	done := make(chan bool)

	go r.CDSAckAck(wildcardRequest, responses, errors, done)
	for {
		select {
		case response := <-responses:
			fmt.Printf("got a response!: %v\n", response)
			return nil
		case error := <-errors:
			fmt.Printf("got an error! %v\n", error)
			return error
		}
	}
}

func (r *Runner) TheRunnerReceivesTheFollowing(resourceType string, resources *godog.DocString) error {
	return godog.ErrPending
}
