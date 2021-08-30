package runner

import (
	"context"
	"errors"
	"fmt"
	"time"
	"reflect"

	"github.com/cucumber/godog"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/ii/xds-test-harness/internal/parser"
	pb "github.com/zachmandeville/tester-prototype/api/adapter"
)

func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a target setup with the following state:$`, r.ATargetSetupWithTheFollowingState)
	ctx.Step(`^the Client sends an initial CDS wildcard request$`, r.ClientSendsAnInitialCDSWildcardRequest)
	ctx.Step(`^the Client receives the following version and clusters:$`, r.ClientReceivesTheFollowingVersionAndClusters)
	ctx.Step(`^cluster "([^"]*)" is updated to version "([^"]*)" after Client subscribed to CDS$`, r.ClusterIsUpdatedToVersionAfterClientSubscribedToCDS)
}

func (r *Runner) ATargetSetupWithTheFollowingState(state *godog.DocString) error {
	snapshot, err := parser.YamlToSnapshot(r.NodeID, state.Content)
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
	r.Cache.Snapshot = snapshot
	return err
}

func (r *Runner) ClientSendsAnInitialCDSWildcardRequest() error {
	requests := make(chan *discovery.DiscoveryRequest, 1)
	responses := make(chan *discovery.DiscoveryResponse, 1)
	errors := make(chan error, 1)
	done := make(chan bool, 1)

	go r.CDSAckAck(requests, responses, errors, done)

	request := NewWildcardCDSRequest(r.NodeID)
	requests <- request

	for {
		select {
		case res := <-responses:
			ackRequest, _ := NewCDSAckRequestFromResponse(r.NodeID, res)
			requests <- ackRequest
			r.Cache.Response = res
			close(requests)
		case err := <-errors:
			err = fmt.Errorf("Error while receiving responses from CDS: %v", err)
			close(requests)
			return err
		case <-done:
			return nil
		}
	}
}


func (r *Runner) ClientReceivesTheFollowingVersionAndClusters(resources *godog.DocString) error {
	expected, err := parser.YamlToSnapshot(r.NodeID, resources.Content)
	if err != nil {
		fmt.Printf("error parsing snapshot: %v", err)
	}

	expectedVersion := expected.GetVersion()
	expectedClusters := []string{}
	for _, cluster := range expected.Clusters.Items {
		expectedClusters = append(expectedClusters, cluster.GetName())
	}

	response, err := parser.ParseDiscoveryResponse(r.Cache.Response)
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

func (r *Runner) ClusterIsUpdatedToVersionAfterClientSubscribedToCDS(cluster string, version string) error {
	requests := make(chan *discovery.DiscoveryRequest, 1)
	responses := make(chan *discovery.DiscoveryResponse, 1)
	errors := make(chan error, 1)
	done := make(chan bool, 1)

	go r.CDSAckAck(requests, responses, errors, done)

	request := NewWildcardCDSRequest("test-id")
	requests <- request

	adapter := pb.NewAdapterClient(r.Adapter.Conn)

	for {
		select {
		case res := <-responses:
			ackRequest, _ := NewCDSAckRequestFromResponse("test-id", res)
			requests <- ackRequest
			r.Cache.Response = res
			if res.VersionInfo != version {
				newState := *r.Cache.Snapshot
				newState.Version = "2"
				for _, cluster := range newState.Clusters.Items {
					cluster.ConnectTimeout = map[string]int32{"seconds": 10}
				}
				_, err := adapter.SetState(context.Background(), &newState)
				if err != nil {
					fmt.Println("ERROR SETTING NEW STATE!")
				}
			}
			if res.VersionInfo == version {
				time.Sleep(2 * time.Second)
				close(requests)
			}
		case err := <-errors:
			err = fmt.Errorf("Error while receiving responses from CDS: %v", err)
			close(requests)
			return err
		case <-done:
			return nil
		default:
		}
	}
}
