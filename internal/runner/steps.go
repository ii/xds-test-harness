package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	parser "github.com/ii/xds-test-harness/internal/parser"
	"github.com/rs/zerolog/log"
	pb "github.com/ii/xds-test-harness/api/adapter"
)

func itemInSlice(item string, slice []string) bool {
	for _, sliceItem := range slice {
		if item == sliceItem {
			return true
		}
	}
	return false
}

func versionsMatch(expected string, actual string) bool {
	return expected == actual
}

func resourcesMatch(expected []string, actual []string) bool {
	// Compare the resources in a discovery response to the ones we expect.
	// It is valid for the response to give more resources than subscribed to,
	// which is why we are not checking the equality of the two slices, only that
	// all of expected is contained in actual.
	for _, ec := range expected {
		if match := itemInSlice(ec, actual); match == false  {
			return false
		}
	}
	return true
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
	// We initiate a subscription with a request, and the testing client is set
	// to ACK every response. Because of this, here should always be one more
	// request than response, being that first subscribing request. If there are
	// more responses than requests, it strongly indicates the server responded
	// to every ack, including the last one, which is not conformant.
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
	resources := []string{}
	if service == "CDS" {
		r.ClientSubscribesToCDS(resources)
	}
	if service == "LDS" {
		r.ClientSubscribesToLDS(resources)
	}
	return nil
}

func(r *Runner) ClientSubscribesToCDS (resources []string) error {
	r.CDS.Req = make(chan *discovery.DiscoveryRequest, 1)
	r.CDS.Res = make(chan *discovery.DiscoveryResponse, 1)
	r.CDS.Err = make(chan error, 1)
	r.CDS.Done = make(chan bool, 1)
	r.CDS.Cache.InitResource = resources

	typeURL := "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	request := r.NewRequest(r.CDS.Cache.InitResource, typeURL)

	log.Debug().
		Msgf("Sending subscribing request: %v\n", request)
	go r.CDSStream()
	go r.Ack(request, r.CDS)
	return nil

}

func (r *Runner) ClientSubscribesToLDS (resources []string) error {
	r.LDS.Req = make(chan *discovery.DiscoveryRequest, 1)
	r.LDS.Res = make(chan *discovery.DiscoveryResponse, 1)
	r.LDS.Err = make(chan error, 1)
	r.LDS.Done = make(chan bool, 1)
	r.LDS.Cache.InitResource = resources

	typeURL := "type.googleapis.com/envoy.config.listener.v3.Listener"
	request := r.NewRequest(r.LDS.Cache.InitResource, typeURL)

	log.Debug().
		Msgf("Sending subscribing request: %v\n", request)
	go r.LDSStream()
	go r.Ack(request, r.LDS)
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
					if versionsMatch(version, actual.Version) && resourcesMatch(expectedResources, actual.Resources) {
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


func (r *Runner) ResourceIsAddedToServiceWithVersion(resource, service, version string) error {
	log.Debug().
		Msgf("Adding %v to %v service", resource, service)

	snapshot := r.Cache.StartState
	snapshot.Version = version

	if service == "LDS" {
		listeners := snapshot.GetListeners()
		newListener := &pb.Listeners_Listener{
			Name:    resource,
			Address: parser.RandomAddress(),
		}
		listeners.Items = append(listeners.Items, newListener)
		snapshot.Listeners = listeners
	}

	if service == "CDS" {
		clusters := snapshot.GetClusters()
		newCluster := &pb.Clusters_Cluster{
			Name:           resource,
			ConnectTimeout: map[string]int32{"seconds": 5},
		}
		clusters.Items = append(clusters.Items, newCluster)
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

func (r *Runner) ClientSubscribesToASubsetOfResourcesForService(subset, service string) error {
	resources := strings.Split(subset, ",")
	if service == "CDS" {
		r.ClientSubscribesToCDS(resources)
	}
	if service == "LDS" {
		r.ClientSubscribesToLDS(resources)
	}
	return nil
}

func (r *Runner) ClientUpdatesSubscriptionToAResourceForServiceWithVersion(resource, service,version string) error {
	var stream *Service
	var typeURL string

	if service == "LDS" {
		typeURL = "type.googleapis.com/envoy.config.listener.v3.Listener"
		stream = r.LDS
	}

	if service == "CDS" {
		typeURL = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
		stream = r.CDS
	}

	request := &discovery.DiscoveryRequest{
		VersionInfo:   version,
		ResourceNames: []string{resource},
		TypeUrl:       typeURL,
	}
	log.Debug().Msgf("Sending Request: %v", request)
	stream.Req <- request
	return nil
}

func (r *Runner) ClientUnsubcribesFromAllResourcesForService(service string) error {
	var stream *Service
	var typeURL string

	if service == "LDS" {
		typeURL = "type.googleapis.com/envoy.config.listener.v3.Listener"
		stream = r.LDS
	}

	if service == "CDS" {
		typeURL = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
		stream = r.CDS
	}

	version := r.Cache.StartState.Version

	request := &discovery.DiscoveryRequest{
		VersionInfo:   version,
		TypeUrl:       typeURL,
	}
	time.Sleep(3 * time.Second)
	log.Debug().Msgf("Sending unsubscribe request: %v", request)
	stream.Req <- request
	time.Sleep(3 * time.Second)
	return nil
}

func (r *Runner) ClientDoesNotReceiveAnyMessageFromService(service string) error {
	var stream *Service

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
			return err
		default:
			if len(stream.Cache.Responses) > 0 {
				for _, response := range stream.Cache.Responses {
					actual, err := parser.ParseDiscoveryResponseV2(response)
					if err != nil {
						log.Error().Err(err).Msg("can't parse discovery response ")
						return err
					}
					err = errors.New("Received a response when we expected no response")
					log.Err(err).Msgf("Response: %v",actual)
					return err
				}
			}
		}
	}
}


func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
    ctx.Step(`^a target setup with "([^"]*)", "([^"]*)", and "([^"]*)"$`, r.ATargetSetupWithServiceResourcesAndVersion)
	ctx.Step(`^the Client does a wildcard subscription to "([^"]*)"$`, r.TheClientDoesAWildcardSubscriptionToService)
    ctx.Step(`^the Client subscribes to a "([^"]*)" for "([^"]*)"$`, r.ClientSubscribesToASubsetOfResourcesForService)
    ctx.Step(`^the Client receives the "([^"]*)" and "([^"]*)" for "([^"]*)"$`, r.TheClientReceivesCorrectResourcesAndVersionForService)
	ctx.Step(`^the Client sends an ACK to which the "([^"]*)" does not respond$`, r.TheClientSendsAnACKToWhichTheDoesNotRespond)
    ctx.Step(`^a "([^"]*)" of the "([^"]*)" is updated to the "([^"]*)"$`, r.ResourceOfTheServiceIsUpdatedToNextVersion)
	ctx.Step(`^the client receives the "([^"]*)" and "([^"]*)" for "([^"]*)"$`, r.TheClientReceivesCorrectResourcesAndVersionForService)
    ctx.Step(`^a "([^"]*)" is added to the "([^"]*)" with "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
    ctx.Step(`^the Client updates subscription to a "([^"]*)" of "([^"]*)" with "([^"]*)"$`, r.ClientUpdatesSubscriptionToAResourceForServiceWithVersion)
	ctx.Step(`^the Client does not receive any message from "([^"]*)"$`, r.ClientDoesNotReceiveAnyMessageFromService)
	ctx.Step(`^the Client unsubcribes from all resources for "([^"]*)"$`, r.ClientUnsubcribesFromAllResourcesForService)
}
