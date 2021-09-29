package runner

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cucumber/godog"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	parser "github.com/ii/xds-test-harness/internal/parser"
	"github.com/rs/zerolog/log"
	pb "github.com/ii/xds-test-harness/api/adapter"
)

func sortCompare(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)

	for i, str := range a {
		if str != b[i] {
			return false
		}
	}
	return true
}

func versionsMatch(expected *pb.Snapshot, actual *parser.DiscoveryResponse) bool {
	return expected.GetVersion() == actual.VersionInfo
}

func clustersMatch(expected *pb.Snapshot, actual *parser.DiscoveryResponse) bool {
	expectedClusters := []string{}
	for _, ec := range expected.Clusters.Items {
		expectedClusters = append(expectedClusters, ec.GetName())
	}
	actualClusters := []string{}
	for _, ac := range actual.Resources {
		actualClusters = append(actualClusters, ac.Name)
	}
	return sortCompare(expectedClusters, actualClusters)
}

func (r *Runner) ClientSubscribesToWildcardCDS() error {
	r.CDS.Req = make(chan *discovery.DiscoveryRequest, 1)
	r.CDS.Res = make(chan *discovery.DiscoveryResponse, 1)
	r.CDS.Err = make(chan error, 1)
	r.CDS.Done = make(chan bool, 1)
	r.CDS.Cache.InitResource = []string{}

	typeURL := "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	request := r.NewRequest(r.CDS.Cache.InitResource, typeURL)

	go r.CDSStream()
	go r.Ack(request, r.CDS)
	return nil
}

func (r *Runner) ClientSubscribesToWildcardLDS() {
	r.LDS.Req = make(chan *discovery.DiscoveryRequest, 1)
	r.LDS.Res = make(chan *discovery.DiscoveryResponse, 1)
	r.LDS.Err = make(chan error, 1)
	r.LDS.Done = make(chan bool, 1)
	r.LDS.Cache.InitResource = []string{}

	typeURL := "type.googleapis.com/envoy.config.listener.v3.Listener"

	request := r.NewRequest(r.LDS.Cache.InitResource, typeURL)

	go r.LDSStream()
	go r.Ack(request, r.LDS)
}

func (r *Runner) TheClientSendsAnACKToWhichTheServerDoesNotRespond() error {
	r.CDS.Done <- true
	// give some time for the final messages to come through, if there's any lingering responses.
	time.Sleep(3 * time.Second)
	log.Debug().
		Msgf("Request Count: %v Response Count: %v", len(r.CDS.Cache.Requests), len(r.CDS.Cache.Responses))
	if len(r.CDS.Cache.Requests) <= len(r.CDS.Cache.Responses) {
		err := errors.New("There are more responses than requests.  This indicates the server responded to the last ack")
		log.Err(err).
			Msgf("Requests:%v, Responses: \v", r.CDS.Cache.Requests, r.CDS.Cache.Responses)
		return err
	}
	return nil
}


func (r *Runner) ATargetSetupWithServiceResourcesAndVersion(service, resources, version string) error {
	snapshot := &pb.Snapshot{
		Node:      r.NodeID,
		Version:   fmt.Sprint(version),
		Clusters: &pb.Clusters{},
	}

	if service == "LDS" {
		listeners := parser.ToListeners(resources)
		snapshot.Listeners = listeners
	}
	if service == "CDS" {
		clusters := parser.ToClusters(resources)
		snapshot.Clusters = clusters
	}

	c := pb.NewAdapterClient(r.Adapter.Conn)

	_, err := c.SetState(context.Background(), snapshot)
	if err != nil {
		msg := "Cannot set target with given state"
		log.Error().
			Err(err).
			Msg(msg)
		return errors.New(msg)
	}

	r.Cache.StartState = snapshot
	return nil
}

func (r *Runner) TheClientDoesAWildcardSubscriptionToService(service string) error {
	if service == "CDS" {
		r.ClientSubscribesToWildcardCDS()
	}
	if service == "LDS" {
		r.ClientSubscribesToWildcardLDS()
	}
	return nil
}

func (r *Runner) TheClientReceivesCorrectResourcesAndVersionForService(resources, version, service string) error {
	var stream *Service
	expectedResources := strings.Split(resources, ",")

	if service == "CDS" {
		stream = r.CDS
	}
	if service == "LDS" {
		stream = r.LDS
	}

	for {
		select {
		case err := <- stream.Err:
			log.Err(err).Msg("From our step")
			return errors.New("Could not find expected response within grace period of 10 seconds.")
		default:
			if len(stream.Cache.Responses) > 0 {
				for _, response := range stream.Cache.Responses {
				    actual, err := parser.ParseDiscoveryResponseV2(response)
					if err != nil {
						log.Error().Err(err).Msg("can't parse discovery response ")
						return err
					}
					if actual.Version == version && sortCompare(expectedResources, actual.Resources) {
						return nil
					}
				}
			}
		}
	}
}

func (r *Runner) TheClientSendsAnACKToWhichTheDoesNotRespond(service string) error {
	var stream *Service
	if service == "CDS" {
		stream = r.CDS
	}
	if service == "LDS" {
		stream = r.LDS
	}
	stream.Done <- true

	// give some time for the final messages to come through, if there's any lingering responses.
	time.Sleep(3 * time.Second)
	log.Debug().
		Msgf("Request Count: %v Response Count: %v", len(stream.Cache.Requests), len(stream.Cache.Responses))
	if len(stream.Cache.Requests) <= len(stream.Cache.Responses) {
		err := errors.New("There are more responses than requests.  This indicates the server responded to the last ack")
		log.Err(err).
			Msgf("Requests:%v, Responses: \v", stream.Cache.Requests, stream.Cache.Responses)
		return err
	}
	return nil
}

func (r *Runner) ResourceOfTheServiceIsUpdatedToNextVersion(resource, service, version string) error {
	log.Debug().
		Msgf("Updating target state for %v resource %v", service, resource)

	snapshot := r.Cache.StartState
	snapshot.Version = version

	if service == "LDS" {
		listeners := snapshot.GetListeners()
		for _, listener := range listeners.Items {
			if listener.Name == resource {
				listener.Address = parser.RandomAddress()
			}
		}
		snapshot.Listeners = listeners
	}

	if service == "CDS" {
		clusters := snapshot.GetClusters()
		for _, cluster := range clusters.Items {
			if cluster.Name == resource {
				cluster.ConnectTimeout = map[string]int32{"seconds": 10}
			}
		}
		snapshot.Clusters = clusters
	}

	c := pb.NewAdapterClient(r.Adapter.Conn)

	_, err := c.UpdateState(context.Background(), snapshot)
	if err != nil {
		msg := "Cannot update target with given state"
		log.Error().
			Err(err).
			Msg(msg)
		return errors.New(msg)
	}

	r.Cache.StateSnapshots = append(r.Cache.StateSnapshots, snapshot)

	return nil
}

func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
    ctx.Step(`^a target setup with "([^"]*)", "([^"]*)", and "([^"]*)"$`, r.ATargetSetupWithServiceResourcesAndVersion)
	ctx.Step(`^the Client does a wildcard subscription to "([^"]*)"$`, r.TheClientDoesAWildcardSubscriptionToService)
    ctx.Step(`^the Client receives the "([^"]*)" and "([^"]*)" for "([^"]*)"$`, r.TheClientReceivesCorrectResourcesAndVersionForService)
	ctx.Step(`^the Client sends an ACK to which the "([^"]*)" does not respond$`, r.TheClientSendsAnACKToWhichTheDoesNotRespond)
    ctx.Step(`^a "([^"]*)" of the "([^"]*)" is updated to the "([^"]*)"$`, r.ResourceOfTheServiceIsUpdatedToNextVersion)
	ctx.Step(`^the client receives the "([^"]*)" and "([^"]*)" for "([^"]*)"$`, r.TheClientReceivesCorrectResourcesAndVersionForService)
}
