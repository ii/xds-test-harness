package runner

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/cucumber/godog"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/ii/xds-test-harness/internal/parser"
	pb "github.com/zachmandeville/tester-prototype/api/adapter"
)


func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a target setup with the following state:$`, r.ATargetSetupWithTheFollowingState)
	ctx.Step(`^the Runner receives the following clusters:$`, r.TheRunnerReceivesTheFollowingClusters)
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

	requests := make(chan *discovery.DiscoveryRequest, 1)
	responses := make(chan *discovery.DiscoveryResponse, 1)
	errors := make(chan error, 1)
	done := make(chan bool, 1)

	go r.CDSAckAck(requests, responses, errors, done)

	request := NewWildcardCDSRequest(nodeID)
	requests <- request

	for {
		select {
		case res := <- responses:
			ackRequest, _ := NewCDSAckRequestFromResponse(nodeID, res)
			requests <- ackRequest
			r.Results.Response = res
			close(requests)
		case err:= <-errors:
			err = fmt.Errorf("Error while receiving responses from CDS: %v", err)
			close(requests)
			return err
		case <-done:
			return nil
		}
	}
}

func (r *Runner) TheRunnerReceivesTheFollowingClusters(resources *godog.DocString) error {
	expected, err := parser.YamlToSnapshot(resources.Content)
	if err != nil {
		fmt.Printf("error parsing snapshot: %v", err)
	}

	expectedVersion := expected.GetVersion()
	expectedClusters := []string{}
	for _, cluster := range expected.Clusters.Items {
		expectedClusters = append(expectedClusters, cluster.GetName())
	}

	response, err := parser.ParseDiscoveryResponse(r.Results.Response)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
	}

	actualVersion := response.VersionInfo
	if expectedVersion != actualVersion {
		err := fmt.Errorf("expected version doesn't match actual version: %v", err)
		return err
	}
	actualClusters := []string{}
	for _, cluster := range response.Resources {
		actualClusters = append(actualClusters, cluster.Name)
	}

	clustersMatch := reflect.DeepEqual(expectedClusters, actualClusters)
	if !clustersMatch {
		err := fmt.Errorf("Clusters don't match.\nexpected:%v\nactual:%v\n", expectedClusters, actualClusters)
		return err
	}
	return nil
}
