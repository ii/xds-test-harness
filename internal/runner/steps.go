package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/ii/xds-test-harness/api/adapter"
	pb "github.com/ii/xds-test-harness/api/adapter"
	parser "github.com/ii/xds-test-harness/internal/parser"
	"github.com/rs/zerolog/log"
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
	log.Debug().Msgf("resources to match...expected: %v, actual: %v\n", expected, actual)
	// Compare the resources in a discovery response to the ones we expect.
	// It is valid for the response to give more resources than subscribed to,
	// which is why we are not checking the equality of the two slices, only that
	// all of expected is contained in actual.
	for _, ec := range expected {
		if match := itemInSlice(ec, actual); match == false {
			return false
		}
	}
	return true
}

func (r *Runner) ATargetSetupWithServiceResourcesAndVersion(service, resources, version string) error {
	snapshot := &pb.Snapshot{
		Node:    r.NodeID,
		Version: fmt.Sprint(version),
	}
	resourceNames := strings.Split(resources, ",")

	//Set endpoints
	endpoints := parser.ToEndpoints(resourceNames)
	snapshot.Endpoints = endpoints

	//Set clusters
	clusters := parser.ToClusters(resourceNames)
	snapshot.Clusters = clusters

	//Set Routes
	routes := parser.ToRoutes(resourceNames)
	snapshot.Routes = routes

	//Set listeners
	listeners := parser.ToListeners(resourceNames)
	snapshot.Listeners = listeners

	//Set runtimes
	runtimes := parser.ToRuntimes(resourceNames)
	snapshot.Runtimes = runtimes

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
	r.ClientSubscribesToServiceForResources(service, resources)
	return nil
}

func (r *Runner) ClientSubscribesToASubsetOfResourcesForService(subset, service string) error {
	resources := strings.Split(subset, ",")
	r.ClientSubscribesToServiceForResources(service, resources)
	return nil
}

func (r *Runner) ClientSubscribesToServiceForResources(srv string, resources []string) error {
	builder := getBuilder(srv)
	builder.openChannels()
	builder.setInitResources(resources)
	err := builder.setStream(r.Target.Conn)
	if err != nil {
		return err
	}
	r.Service = builder.getService()

	request := r.NewRequest(r.Service.Cache.InitResource, r.Service.TypeURL)

	log.Debug().
		Msgf("Sending subscribing request: %v\n", request)
	go r.Stream(r.Service)
	go r.Ack(request, r.Service)
	return nil
}

func (r *Runner) TheClientReceivesCorrectResourcesAndVersion(resources, version string) error {
	expectedResources := strings.Split(resources, ",")
	stream := r.Service
	actualResources := []string{}

	for {
		select {
		case err := <-stream.Channels.Err:
			log.Err(err).Msg("From our step")
			return errors.New("Could not find expected response within grace period of 10 seconds.")
		default:
			if len(stream.Cache.Responses) > 0 {
				for _, response := range stream.Cache.Responses {
					actual, err := parser.ParseDiscoveryResponse(response)
					if err != nil {
						log.Error().Err(err).Msg("can't parse discovery response ")
						return err
					}
					if !versionsMatch(version, actual.Version) {
						continue
					}
					if stream.Name == "RDS" || stream.Name == "EDS" { // EDS & RDS resources can come from multiple responses.
						actualResources = append(actualResources, actual.Resources...)
					} else {
						actualResources = actual.Resources
					}
					if !resourcesMatch(expectedResources, actualResources) {
						continue
					}
					return nil
				}
			}
		}
	}
}

func (r *Runner) theClientReceivesOnlyTheCorrectResourceAndVersion(resource, version string) error {
	stream := r.Service
	log.Debug().Msgf("Resource: %v, version: %v", resource, version)

	for {
		select {
		case err := <-stream.Channels.Err:
			log.Err(err).Msg("From our step")
			return errors.New("Could not find expected response within grace period of 10 seconds.")
		default:
			if len(stream.Cache.Responses) > 0 {
				for _, response := range stream.Cache.Responses {
					actual, err := parser.ParseDiscoveryResponse(response)
					if err != nil {
						log.Error().Err(err).Msg("can't parse discovery response ")
						return err
					}
					if !versionsMatch(version, actual.Version) {
						continue
					}
					// we set our subscription to a single resource, and the services should only send a single resource.
					// If the resources slice is empty or more than one, it is incorrect and we can continue.
					// (this fn is not designed for LDS or CDS tests)
					if len(actual.Resources) != 1 {
						continue
					}
					if !resourcesMatch([]string{resource}, actual.Resources) {
						log.Debug().Msgf("resources don't match: %v", actual.Resources)
						continue
					}
					return nil
				}
			}
		}
	}
}

func (r *Runner) TheClientSendsAnACKToWhichTheDoesNotRespond(service string) error {
	stream := r.Service
	stream.Channels.Done <- true

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
	snapshot := &adapter.Snapshot{
		Version: version,
	}
	startState := r.Cache.StartState
	snapshot.Node = startState.Node
	snapshot.Endpoints = startState.Endpoints
	snapshot.Clusters = startState.Clusters
	snapshot.Routes = startState.Routes
	snapshot.Listeners = startState.Listeners
	snapshot.Runtimes = startState.Runtimes
	snapshot.Secrets = startState.Secrets

	// TODO I am uncertain if we need to update the contents of any
	// resource? Is updating just the version enough to trigger
	// a subscription update? If so, we can remove the majority of this
	// function body.
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

	//Set endpoints
	endpoints := snapshot.GetEndpoints()
	endpoints.Items = append(endpoints.Items, &pb.Endpoint{
		Name:    resource,
		Cluster: resource,
		Address: parser.RandomAddress(),
	})
	snapshot.Endpoints = endpoints

	//Set clusters
	clusters := snapshot.GetClusters()
	clusters.Items = append(clusters.Items, &pb.Cluster{
		Name:           resource,
		ConnectTimeout: map[string]int32{"seconds": 5},
	})
	snapshot.Clusters = clusters

	//Set Routes
	routes := snapshot.GetRoutes()
	routes.Items = append(routes.Items, &pb.Route{
		Name: resource,
	})
	snapshot.Routes = routes

	//Set listeners
	listeners := snapshot.GetListeners()
	listeners.Items = append(listeners.Items, &pb.Listener{
		Name:    resource,
		Address: parser.RandomAddress(),
	})

	//Set runtimes
	runtimes := snapshot.GetRuntimes()
	runtimes.Items = append(runtimes.Items, &pb.Runtime{
		Name: resource,
	})
	snapshot.Runtimes = runtimes

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

func (r *Runner) ClientUpdatesSubscriptionToAResourceForServiceWithVersion(resource, service, version string) error {
	request := &discovery.DiscoveryRequest{
		VersionInfo:   version,
		ResourceNames: []string{resource},
		TypeUrl:       r.Service.TypeURL,
	}
	// small hack as i build this out, to ensure Acking our last response happens before we update subscription
	time.Sleep(2 * time.Second)
	log.Debug().
		Msgf("Sending Request To Update Subscription: %v", request)
	r.Service.Channels.Req <- request
	return nil
}

func (r *Runner) ClientUnsubscribesFromAllResourcesForService(service string) error {
	version := r.Cache.StartState.Version

	request := &discovery.DiscoveryRequest{
		VersionInfo:   version,
		ResourceNames: []string{""},
		TypeUrl:       r.Service.TypeURL,
	}
	time.Sleep(3 * time.Second)
	log.Debug().
		Msgf("Sending unsubscribe request: %v", request)
	r.Service.Channels.Req <- request
	time.Sleep(3 * time.Second)
	return nil
}

func (r *Runner) ClientDoesNotReceiveAnyMessageFromService(service string) error {
	for {
		select {
		case err := <-r.Service.Channels.Err:
			log.Err(err).Msg("From our step")
			return err
		default:
			if len(r.Service.Cache.Responses) > 0 {
				for _, response := range r.Service.Cache.Responses {
					actual, err := parser.ParseDiscoveryResponse(response)
					if err != nil {
						log.Error().
							Err(err).
							Msg("can't parse discovery response ")
						return err
					}
					currentState := r.Cache.StateSnapshots[len(r.Cache.StateSnapshots)-1]
					if actual.Version != currentState.Version {
						continue
					}
					err = errors.New("Received a response when we expected no response")
					log.Err(err).
						Msgf("Response: %v", actual)
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
	ctx.Step(`^the Client receives the "([^"]*)" and "([^"]*)"$`, r.TheClientReceivesCorrectResourcesAndVersion)
	ctx.Step(`^the Client receives only the "([^"]*)" and "([^"]*)"$`, r.theClientReceivesOnlyTheCorrectResourceAndVersion)
	ctx.Step(`^the Client does not receive any message from "([^"]*)"$`, r.ClientDoesNotReceiveAnyMessageFromService)
	ctx.Step(`^the Client sends an ACK to which the "([^"]*)" does not respond$`, r.TheClientSendsAnACKToWhichTheDoesNotRespond)
	ctx.Step(`^a "([^"]*)" of the "([^"]*)" is updated to the "([^"]*)"$`, r.ResourceOfTheServiceIsUpdatedToNextVersion)
	ctx.Step(`^a "([^"]*)" is added to the "([^"]*)" with "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
	ctx.Step(`^the Client updates subscription to a "([^"]*)" of "([^"]*)" with "([^"]*)"$`, r.ClientUpdatesSubscriptionToAResourceForServiceWithVersion)
	ctx.Step(`^the Client unsubscribes from all resources for "([^"]*)"$`, r.ClientUnsubscribesFromAllResourcesForService)
}
