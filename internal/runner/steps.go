package runner

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cucumber/godog"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/ii/xds-test-harness/api/adapter"
	pb "github.com/ii/xds-test-harness/api/adapter"
	parser "github.com/ii/xds-test-harness/internal/parser"
	utils "github.com/ii/xds-test-harness/internal/utils"
	"github.com/rs/zerolog/log"
)

func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a target setup with service "([^"]*)", resources "([^"]*)", and starting version "([^"]*)"$`, r.TargetSetupWithServiceResourcesAndVersion)
	ctx.Step(`^the resources "([^"]*)" of the "([^"]*)" is updated to version "([^"]*)"$`, r.ResourceOfTheServiceIsUpdatedToNextVersion)
	ctx.Step(`^the Client does a wildcard subscription to "([^"]*)"$`, r.ClientDoesAWildcardSubscriptionToService)
	ctx.Step(`^the Client subscribes to a subset of resources,"([^"]*)", for "([^"]*)"$`, r.ClientSubscribesToASubsetOfResourcesForService)
	ctx.Step(`^the Client subscribes to resources "([^"]*)" for "([^"]*)"$`, r.ClientSubscribesToASubsetOfResourcesForService)
	ctx.Step(`^the Client receives the resources "([^"]*)" and version "([^"]*)" for "([^"]*)"$`, r.ClientReceivesResourcesAndVersionForService)
	ctx.Step(`^the Client receives only the resource "([^"]*)" and version "([^"]*)"$`, r.ClientReceivesOnlyTheCorrectResourceAndVersion)
	ctx.Step(`^the Client does not receive any message from "([^"]*)"$`, r.ClientDoesNotReceiveAnyMessageFromService)
	ctx.Step(`^the Client sends an ACK to which the "([^"]*)" does not respond$`, r.TheServiceNeverRespondsMoreThanNecessary)
	ctx.Step(`^the resource "([^"]*)" is added to the "([^"]*)" with version "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
	ctx.Step(`^a resource "([^"]*)" is added to the "([^"]*)" with version "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
	ctx.Step(`^the resources "([^"]*)" are added to the "([^"]*)" with version "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
	ctx.Step(`^the Client updates subscription to a resource\("([^"]*)"\) of "([^"]*)" with version "([^"]*)"$`, r.ClientUpdatesSubscriptionToAResourceForServiceWithVersion)
	ctx.Step(`^the Client updates subscription to a "([^"]*)" of "([^"]*)" with "([^"]*)"$`, r.ClientUpdatesSubscriptionToAResourceForServiceWithVersion) // delete?
	ctx.Step(`^the Client unsubscribes from all resources for "([^"]*)"$`, r.ClientUnsubscribesFromAllResourcesForService)
	ctx.Step(`^the Client receives the "([^"]*)" and "([^"]*)" for "([^"]*)"$`, r.ClientReceivesResourcesAndVersionForService)
	ctx.Step(`^the service never responds more than necessary$`, r.TheServiceNeverRespondsMoreThanNecessary)
	ctx.Step(`^the Client receives only the resource "([^"]*)" and version "([^"]*)" for service$`, r.ClientReceivesOnlyTheResourceAndVersionForService)
	ctx.Step(`^the resource "([^"]*)" of service "([^"]*)" is updated to version "([^"]*)"$`, r.ResourceOfServiceIsUpdatedToVersion)
}

// Creates a snapshot to be sent, via the adapter, to the target implementation,
// setting the state for the rest of the steps.
func (r *Runner) TargetSetupWithServiceResourcesAndVersion(service, resources, version string) error {
	snapshot := &pb.Snapshot{
		Node:    r.NodeID,
		Version: fmt.Sprint(version),
	}
	resourceNames := strings.Split(resources, ",")

	//Set endpoints
	snapshot.Endpoints = parser.ToEndpoints(resourceNames)

	//Set clusters
	snapshot.Clusters = parser.ToClusters(resourceNames)

	//Set Routes
	snapshot.Routes = parser.ToRoutes(resourceNames)

	//Set listeners
	snapshot.Listeners = parser.ToListeners(resourceNames)

	//Set runtimes
	snapshot.Runtimes = parser.ToRuntimes(resourceNames)

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

// Uses existing snapshot to build new version.
// All resources are updated to next version for convenience.
// nothing changes about the resources themselves, but the new version
// should still trigger a response.
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

// Wrapper to start stream, without resources, for given service
func (r *Runner) ClientDoesAWildcardSubscriptionToService(service string) error {
	resources := []string{}
	r.ClientSubscribesToServiceForResources(service, resources)
	return nil
}

// Wrapper to start stream, with given resources for given service
func (r *Runner) ClientSubscribesToASubsetOfResourcesForService(subset, service string) error {
	resources := strings.Split(subset, ",")
	if r.Incremental {
		r.Delta_ClientSubscribesToServiceForResources(service, resources)
	} else {
		r.ClientSubscribesToServiceForResources(service, resources)
	}
	return nil
}

// Takes service and creates a runner.Service with a fresh xDS stream
// for the given service. This is the heart of a test, as it sets up
// the request/response loops that verify the service is working properly.
func (r *Runner) ClientSubscribesToServiceForResources(srv string, resources []string) error {
	typeURL, err := parser.ServiceToTypeURL(srv)
	if err != nil {
		return err
	}
	// check if we are updating existing stream or starting a new one.
	if r.Service.Stream != nil {
		request := newRequest(resources, typeURL, r.NodeID)
		r.Service.Channels.Req <- request
		log.Debug().
			Msgf("Sent new subscribing request: %v\n", request)
		return nil
	} else {
		var builder serviceBuilder
		if r.Aggregated {
			builder = getBuilder("ADS")
		} else {
			builder = getBuilder(srv)
		}
		builder.openChannels()
		builder.setInitResources(resources)
		err := builder.setStreams(r.Target.Conn)
		if err != nil {
			return err
		}
		r.Service = builder.getService(srv)

		request := newRequest(r.Service.Cache.InitResource, typeURL, r.NodeID)
		r.SubscribeRequest = request

		log.Debug().
			Msgf("Sending first subscribing request: %v\n", request)
		go r.Stream(r.Service)
		go r.Ack(r.Service)
		return nil
	}
}

func (r *Runner) Delta_ClientSubscribesToServiceForResources(srv string, resources []string) error {
	typeURL, err := parser.ServiceToTypeURL(srv)
	if err != nil {
		return err
	}
	// check if we are updating existing stream or starting a new one.
	if r.Service.Delta != nil {
		request := newDeltaRequest(resources, typeURL, r.NodeID)
		r.Service.Channels.Delta_Req <- request
		log.Debug().
			Msgf("[delta] Sent new subscribing request: %v\n", request)
		return nil
	} else {
		var builder serviceBuilder
		if r.Aggregated {
			builder = getBuilder("ADS")
		} else {
			builder = getBuilder(srv)
		}
		builder.openChannels()
		builder.setInitResources(resources)
		err := builder.setStreams(r.Target.Conn)
		if err != nil {
			return err
		}
		r.Service = builder.getService(srv)

		request := newDeltaRequest(r.Service.Cache.InitResource, typeURL, r.NodeID)
		r.Delta_SubscribeRequest = request

		log.Debug().
			Msgf("Sending first subscribing request: %v\n", request)
		go r.Delta_Stream(r.Service)
		go r.Delta_Ack(r.Service)
		return nil
	}
}

// Loop through the service's response cache until we get the expected response
// or we reach the deadline for the service.
func (r *Runner) ClientReceivesResourcesAndVersionForService(resources, version, service string) error {
	if r.Incremental {
		err := r.DeltaCheckResources(resources, version, service)
		return err
	} else {
		err := r.CheckResources(resources, version, service)
		return err
	}
}

// Loop through the service's response cache until we get the expected response
// or we reach the deadline for the service.
func (r *Runner) CheckResources(resources, version, service string) error {
	expectedResources := strings.Split(resources, ",")
	typeUrl, err := parser.ServiceToTypeURL(service)
	if err != nil {
		err := fmt.Errorf("Cannot determine typeURL for given service: %v\n", service)
		return err
	}
	done := time.After(3 * time.Second)
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return fmt.Errorf("There was an issue when receiving responses", err)
		case <-done:
			match, single_response, err := r.DB.CheckExpectedResources(expectedResources, version, typeUrl)
			if err != nil {
				return err
			}
			if !match {
				return fmt.Errorf("Could not find expected resources in any of the responses")
			}
			if (service == "CDS" || service == "LDS") && !single_response {
				return fmt.Errorf("Found expected resources, but in multiple responses. for CDS or LDS they should be in a single response")
			}
			return nil
		}
	}
}

// Loop through the service's response cache until we get the expected response
// or we reach the deadline for the service.
func (r *Runner) DeltaCheckResources(resources, version, service string) error {
	expectedResources := strings.Split(resources, ",")
	stream := r.Service
	actualResources := []string{}

	typeURL, err := parser.ServiceToTypeURL(service)
	if err != nil {
		err := fmt.Errorf("Cannot determine typeURL for given service: %v\n", service)
		return err
	}
	for {
		select {
		case err := <-stream.Channels.Err:
			// if there isn't ane rror in a response,
			// the error will be passed down from the stream when
			// it reaches its context deadline.
			return fmt.Errorf("Could not find expected response within grace period of 10 seconds. %v", err)
		default:
			if len(stream.Cache.Delta_Responses) > 0 {
				for _, response := range stream.Cache.Delta_Responses {
					resourceNames, err := parser.DeltaResourceNames(response)
					if err != nil {
						return err
					}
					if !reflect.DeepEqual(version, response.SystemVersionInfo) {
						continue
					}
					if !reflect.DeepEqual(typeURL, response.TypeUrl) {
						continue
					}
					actualResources = append(actualResources, resourceNames...)
					if !resourcesMatch(expectedResources, actualResources) {
						continue
					}
					return nil
				}
			}
		}
	}
}

// Loop again, but this time continuing if resource in cache has more than one entry.
// The test is itended for when you update a subscription to now only care about a single resource.
// The response you reeceive should only have a single entry in its resources, otherwise we fail.
// Won't work for LDS/CDS where it is conformant to pass along more than you need.
func (r *Runner) ClientReceivesOnlyTheCorrectResourceAndVersion(resource, version string) error {
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
					resourceNames, err := parser.ResourceNames(response)
					if err != nil {
						return fmt.Errorf("Could not parse resource names from response. cannot validate response. response:%v\nerr: %v", response, err)
					}
					if err != nil {
						log.Error().Err(err).Msg("can't parse discovery response ")
						return err
					}
					if !reflect.DeepEqual(version, response.VersionInfo) {
						continue
					}
					// we set our subscription to a single resource, and the services should only send a single resource.
					// If the resources slice is empty or more than one, it is incorrect and we can continue.
					// (this fn is not designed for LDS or CDS tests)
					if len(resourceNames) != 1 {
						// log.Debug().
						// 	Msgf("Got right version, but too many resources: %v", stream.Cache.Requests)
						continue
					}
					if !resourcesMatch([]string{resource}, resourceNames) {
						log.Debug().Msgf("resources don't match: %v", resourceNames)
						continue
					}
					return nil
				}
			}
		}
	}
}

func (r *Runner) TheServiceNeverRespondsMoreThanNecessary() error {
	r.Service.Channels.Done <- true // send signal to close the channels and service down
	correctAmount, err := r.DB.CheckMoreRequestsThanResponses()
	if err != nil {
		return err
	}
	if !correctAmount {
		return fmt.Errorf("Responses were equal, or more, than requests. This indicates the server responded to the last ACK.")
	}
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
	typeURL, err := parser.ServiceToTypeURL(service)
	if err != nil {
		err := fmt.Errorf("Cannot determine typeURL for given service: %v\n", service)
		return err
	}

	lastResponse := r.Service.Cache.Responses[len(r.Service.Cache.Responses)-1]

	request := &discovery.DiscoveryRequest{
		VersionInfo:   lastResponse.VersionInfo,
		ResourceNames: []string{resource},
		TypeUrl:       typeURL,
		ResponseNonce: lastResponse.Nonce,
	}
	r.SubscribeRequest = request
	log.Debug().
		Msgf("Sending Request To Update Subscription: %v", request)
	r.Service.Channels.Req <- request
	return nil
}

func (r *Runner) ClientUnsubscribesFromAllResourcesForService(service string) error {
	// version := r.Cache.StartState.Version
	typeURL, err := parser.ServiceToTypeURL(service)
	if err != nil {
		err := fmt.Errorf("Cannot determine typeURL for given service: %v\n", service)
		return err
	}

	lastResponse := r.Service.Cache.Responses[len(r.Service.Cache.Responses)-1]

	request := &discovery.DiscoveryRequest{
		VersionInfo:   lastResponse.VersionInfo,
		ResourceNames: []string{""},
		TypeUrl:       typeURL,
		ResponseNonce: lastResponse.Nonce,
	}
	r.SubscribeRequest = request
	log.Debug().
		Msgf("Sending unsubscribe request: %v", request)
	r.Service.Channels.Req <- request
	log.Debug().Msg("Pausing for 2 seconds, to ensure server receives unsubscribe test.")
	time.Sleep(2 * time.Second)
	log.Debug().Msg("Should now be good to update server")
	return nil
}

func (r *Runner) ClientDoesNotReceiveAnyMessageFromService(service string) error {
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return err
		default:
			if len(r.Service.Cache.Responses) > 0 {
				for _, response := range r.Service.Cache.Responses {
					currentState := r.Cache.StateSnapshots[len(r.Cache.StateSnapshots)-1]
					if response.VersionInfo != currentState.Version {
						continue
					}
					// a matching version with no resources implies that it responded
					// correctly to an unsubscribe request?
					// if len(response.Resources) == 0 {
					// 	return nil
					// }
					err := errors.New("Received a response when we expected no response")
					log.Err(err).
						Msgf("Response: %v", response)
					return err
				}
			}
		}
	}
}

func (r *Runner) ResourceOfServiceIsUpdatedToVersion(resource, service, version string) error {
	return godog.ErrPending
}

func (r *Runner) ClientReceivesOnlyTheResourceAndVersionForService(resource, version string) error {
	return godog.ErrPending
}

func resourcesMatch(expected []string, actual []string) bool {
	// Compare the resources in a discovery response to the ones we expect.
	// It is valid for the response to give more resources than subscribed to,
	// which is why we are not checking the equality of the two slices, only that
	// all of expected is contained in actual.
	for _, ec := range expected {
		if match, _ := utils.ItemInSlice(ec, actual); match == false {
			return false
		}
	}
	return true
}
