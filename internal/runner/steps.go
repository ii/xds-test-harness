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
	ctx.Step(`^the Client receives only the resource "([^"]*)" and version "([^"]*)" for the service "([^"]*)"$`, r.ClientReceivesOnlyTheResourceAndVersionForTheService)
	ctx.Step(`^the Client does not receive any message from "([^"]*)"$`, r.ClientDoesNotReceiveAnyMessageFromService)
	ctx.Step(`^the resources "([^"]*)" and version "([^"]*)" for "([^"]*)" came in a single response$`, r.ResourcesAndVersionForServiceCameInASingleResponse)
	ctx.Step(`^the Client sends an ACK to which the "([^"]*)" does not respond$`, r.TheServiceNeverRespondsMoreThanNecessary)
	ctx.Step(`^the resource "([^"]*)" is added to the "([^"]*)" with version "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
	ctx.Step(`^a resource "([^"]*)" is added to the "([^"]*)" with version "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
	ctx.Step(`^the resources "([^"]*)" are added to the "([^"]*)" with version "([^"]*)"$`, r.ResourceIsAddedToServiceWithVersion)
	ctx.Step(`^the Client updates subscription to a resource\("([^"]*)"\) of "([^"]*)" with version "([^"]*)"$`, r.ClientUpdatesSubscriptionToAResourceForServiceWithVersion)
	ctx.Step(`^the Client updates subscription to a "([^"]*)" of "([^"]*)" with "([^"]*)"$`, r.ClientUpdatesSubscriptionToAResourceForServiceWithVersion) // delete?
	ctx.Step(`^the Client unsubscribes from all resources for "([^"]*)"$`, r.ClientUnsubscribesFromAllResourcesForService)
	ctx.Step(`^the Client receives the "([^"]*)" and "([^"]*)" for "([^"]*)"$`, r.ClientReceivesResourcesAndVersionForService)
	ctx.Step(`^the service never responds more than necessary$`, r.TheServiceNeverRespondsMoreThanNecessary)

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
	log.Debug().Msgf("Updated resource %v to version %v", resource, version)
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
	r.ClientSubscribesToServiceForResources(service, resources)
	return nil
}

// Takes service and creates a runner.Service with a fresh xDS stream
// for the given service. This is the heart of a test, as it sets up
// the request/response loops that verify the service is working properly.
func (r *Runner) ClientSubscribesToServiceForResources(srv string, resources []string) error {
	err, typeUrl := parser.ServiceToTypeURL(srv)
	if err != nil {
		return err
	}

	r.Validate.Resources[typeUrl] = make(map[string]ValidateResource)
	for _, resource := range resources {
		r.Validate.Resources[typeUrl][resource] = ValidateResource{}
	}

	// check if we are updating existing stream or starting a new one.
	if r.Service.Stream != nil {
		request := newRequest(resources, typeUrl, r.NodeID)
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
		err := builder.setStream(r.Target.Conn)
		if err != nil {
			return err
		}
		r.Service = builder.getService(srv)

		request := newRequest(resources, typeUrl, r.NodeID)
		r.SubscribeRequest = request

		log.Debug().
			Msgf("Sending first subscribing request: %v\n", request)
		go r.Stream(r.Service)
		go r.Ack(r.Service)
		return nil
	}
}

// Loop through the service's response cache until we get the expected response
// or we reach the deadline for the service.
func (r *Runner) ClientReceivesResourcesAndVersionForService(resources, version, service string) error {
	expectedResources := strings.Split(resources, ",")
	stream := r.Service

	err, typeUrl := parser.ServiceToTypeURL(service)
	if err != nil {
		err := fmt.Errorf("Cannot determine typeURL for given service: %v\n", service)
		return err
	}
	done := time.After(3 * time.Second)
	for {
		select {
		case err := <-stream.Channels.Err:
			// if there isn't an error in a response,
			// the error will be passed down from the stream when
			// it reaches its context deadline.
			return fmt.Errorf("Could not find expected response within grace period of 10 seconds. %v", err)
		case <-done:
			actualResources := r.Validate.Resources[typeUrl]
			for _, resource := range expectedResources {
				actual, ok := actualResources[resource]
				if !ok {
					return fmt.Errorf("Could not find resource from responses")
				}
				if actual.Version != version {
					return fmt.Errorf("Found resource, but not correct version. Expected: %v, Actual: %v", version, actual.Version)
				}
			}
			return nil
		}
	}
}

// Loop again, but this time continuing if resource in cache has more than one entry.
// The test is itended for when you update a subscription to now only care about a single resource.
// The response you reeceive should only have a single entry in its resources, otherwise we fail.
// Won't work for LDS/CDS where it is conformant to pass along more than you need.
func (r *Runner) ClientReceivesOnlyTheResourceAndVersionForTheService(resource, version, service string) error {
	done := time.After(3 * time.Second)
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return fmt.Errorf("Could not find expected response within grace period of 10 seconds or encountered error: %v.", err)
		case <-done:
			err, typeUrl := parser.ServiceToTypeURL(service)
			if err != nil {
				return fmt.Errorf("Issue converting service to typeUrl, was it written correctly?")
			}
			resources := r.Validate.Resources[typeUrl]
			for name, info := range resources {
				if name != resource || info.Version != version {
					return fmt.Errorf("Received a resource, or a version, we should not have. Expected resource/version: %v/%v. Got: %v/%v",
						resource, version, name, info.Version)
				}
			}
			return nil
		}
	}
}

func (r *Runner) TheServiceNeverRespondsMoreThanNecessary() error {
	stream := r.Service
	stream.Channels.Done <- true

	// give some time for the final messages to come through, if there's any lingering responses.
	time.Sleep(3 * time.Second)
	log.Debug().
		Msgf("Request Count: %v Response Count: %v", r.Validate.RequestCount, r.Validate.ResponseCount)
	if r.Validate.RequestCount <= r.Validate.ResponseCount {
		err := errors.New("There are more responses than requests.  This indicates the server responded to the last ack")
		return err
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
	err, typeUrl := parser.ServiceToTypeURL(service)
	if err != nil {
		err := fmt.Errorf("Cannot determine typeURL for given service: %v\n", service)
		return err
	}

	current := r.Validate.Resources[typeUrl][resource]

	request := &discovery.DiscoveryRequest{
		VersionInfo:   current.Version,
		ResourceNames: []string{resource},
		TypeUrl:       typeUrl,
		ResponseNonce: current.Nonce,
	}

	r.Validate.Resources[typeUrl] = make(map[string]ValidateResource)
	r.Validate.Resources[typeUrl][resource] = ValidateResource{
		Version: current.Version,
		Nonce:   current.Nonce,
	}
	r.SubscribeRequest = request

	log.Debug().Msgf("Sending Request To Update Subscription: %v", request)
	r.Service.Channels.Req <- request
	return nil
}

func (r *Runner) ClientUnsubscribesFromAllResourcesForService(service string) error {
	err, typeURL := parser.ServiceToTypeURL(service)
	if err != nil {
		err := fmt.Errorf("Cannot determine typeURL for given service: %v\n", service)
		return err
	}

	// we just need a nonce to tell the server we are up to dote and this is a new
	// subscription request. Simple way to grab one from the list of 4.
	var lastNonce string
	for _, v := range r.Validate.Resources[typeURL] {
		lastNonce = v.Nonce
	}
	request := &discovery.DiscoveryRequest{
		ResourceNames: []string{""},
		TypeUrl:       typeURL,
		ResponseNonce: lastNonce,
	}
	r.Validate.Resources[typeURL] = make(map[string]ValidateResource)
	r.SubscribeRequest = request
	log.Debug().
		Msgf("Sending unsubscribe request: %v", request)
	r.Service.Channels.Req <- request
	return nil
}

func (r *Runner) ClientDoesNotReceiveAnyMessageFromService(service string) error {
	err, typeUrl := parser.ServiceToTypeURL(service)
	if err != nil {
		return err
	}
	done := time.After(3 * time.Second)
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return err
		case <-done:
			if len(r.Validate.Resources[typeUrl]) > 0 {
				return fmt.Errorf("Resources received is greater than 0: %v", r.Validate.Resources[typeUrl])
			}
			return nil
		}
	}
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

func (r *Runner) ResourcesAndVersionForServiceCameInASingleResponse(resources, version, service string) error {
	err, typeUrl := parser.ServiceToTypeURL(service)
	if err != nil {
		return err
	}
	expected := strings.Split(resources, ",")
	actual := r.Validate.Resources[typeUrl]

	responses := make(map[string]bool)
	for _, resource := range expected {
		info, ok := actual[resource]
		if !ok || info.Version != version {
			return fmt.Errorf("Could not find correct resource in validation struct. This is rare; perhaps recheck how the test was written.")
		}
		responses[info.Nonce] = true
	}
	if len(responses) != 1 {
		return fmt.Errorf("Resources came via multiple responses. This is not conformant for CDS and  LDS tests")
	}
	return nil
}
