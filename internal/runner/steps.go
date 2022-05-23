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
	ctx.Step(`^the Client receives only the resource "([^"]*)" and version "([^"]*)" for service "([^"]*)"$`, r.ClientReceivesOnlyTheResourceAndVersionForService)
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
	ctx.Step(`^the Delta Client receives only the resource "([^"]*)" and version "([^"]*)" for service "([^"]*)"$`, r.DeltaClientReceivesOnlyTheResourceAndVersionForService)
	ctx.Step(`^the resource "([^"]*)" of service "([^"]*)" is updated to version "([^"]*)"$`, r.ResourceOfServiceIsUpdatedToVersion)
	ctx.Step(`^the Delta Client receives notice that resource "([^"]*)" was removed for service "([^"]*)"$`, r.DeltaClientReceivesNoticeThatResourceWasRemovedForService)
	ctx.Step(`^the resource "([^"]*)" is added to the "([^"]*)" at version "([^"]*)"$`, r.ResourceIsAddedToTheServiceAtVersion)
	ctx.Step(`^the resource "([^"]*)" is removed from the "([^"]*)"$`, r.ResourceIsRemovedFromTheService)
	ctx.Step(`^the Client unsubscribes from resource "([^"]*)" for service "([^"]*)"$`, r.ClientUnsubscribesFromResourceForService)
	ctx.Step(`^the Delta client does not receive resource "([^"]*)" of service "([^"]*)" at version "([^"]*)"$`, r.DeltaClientDoesNotReceiveResourceOfServiceAtVersion)
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
	// check if we are updating existing stream or starting a new one.
	typeUrl := parser.ServiceToTypeURL(srv)
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
		builder.setInitResources(resources)
		err := builder.setStreams(r.Target.Conn)
		if err != nil {
			return err
		}
		r.Service = builder.getService(srv)

		request := newRequest(r.Service.Cache.InitResource, typeUrl, r.NodeID)
		r.SubscribeRequest = request

		log.Debug().
			Msgf("Sending first subscribing request: %v\n", request)
		go r.Stream(r.Service)
		go r.Ack(r.Service)
		return nil
	}
}

func (r *Runner) Delta_ClientSubscribesToServiceForResources(srv string, resources []string) error {
	// check if we are updating existing stream or starting a new one.
	typeUrl := parser.ServiceToTypeURL(srv)
	if r.Service.Delta != nil {
		request := newDeltaRequest(resources, typeUrl, r.NodeID)
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

		request := newDeltaRequest(r.Service.Cache.InitResource, typeUrl, r.NodeID)
		r.Delta_SubscribeRequest = request

		log.Debug().
			Msgf("Sending first subscribing request: %v\n", request)
		go r.Delta_Stream(r.Service)
		go r.Delta_Ack(r.Service)
		return nil
	}
}

// Run db query expectin true value, err if anything but.
func (r *Runner) ClientReceivesResourcesAndVersionForService(resources, version, service string) error {
	expectedResources := strings.Split(resources, ",")
	done := time.After(3 * time.Second)
	typeUrl := parser.ServiceToTypeURL(service)
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return fmt.Errorf("There was an issue when receiving responses: %v", err)
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

// Loop again, but this time continuing if resource in cache has more than one entry.
// The test is itended for when you update a subscription to now only care about a single resource.
// The response you reeceive should only have a single entry in its resources, otherwise we fail.
// Won't work for LDS/CDS where it is conformant to pass along more than you need.
func (r *Runner) ClientReceivesOnlyTheResourceAndVersionForService(resource, version, service string) error {
	expectedResources := strings.Split(resource, ",")
	done := time.After(3 * time.Second)
	typeUrl := parser.ServiceToTypeURL(service)
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return fmt.Errorf("Err receiving responses, coult not validate: %v", err)
		case <-done:
			passed, err := r.DB.CheckOnlyExpectedResources(expectedResources, version, typeUrl)
			if err != nil {
				return err
			}
			if !passed {
				return fmt.Errorf("Did not receive only the resource we wanted")
			}
			return nil
		}
	}
}

func (r *Runner) DeltaClientReceivesNoticeThatResourceWasRemovedForService(resource, service string) error {
	resources := strings.Split(resource, ",")
	done := time.After(3 * time.Second)
	typeUrl := parser.ServiceToTypeURL(service)

	for {
		select {
		case err := <-r.Service.Channels.Err:
			return fmt.Errorf("Error receiving responses, could not validate: %v", err)
		case <-done:
			passed, err := r.DB.DeltaCheckRemovedResources(resources, typeUrl)
			if err != nil {
				return err
			}
			if !passed {
				return fmt.Errorf("Did not receive notice the given resources were removed")
			}
			return nil
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
	lastResponse := r.Service.Cache.Responses[len(r.Service.Cache.Responses)-1]
	typeUrl := parser.ServiceToTypeURL(service)
	request := &discovery.DiscoveryRequest{
		VersionInfo:   lastResponse.VersionInfo,
		ResourceNames: []string{resource},
		TypeUrl:       typeUrl,
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
	lastResponse := r.Service.Cache.Responses[len(r.Service.Cache.Responses)-1]
	typeUrl := parser.ServiceToTypeURL(service)

	request := &discovery.DiscoveryRequest{
		VersionInfo:   lastResponse.VersionInfo,
		ResourceNames: []string{""},
		TypeUrl:       typeUrl,
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
	done := time.After(4 * time.Second)
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return err
		case <-done:
			currentState := r.Cache.StateSnapshots[len(r.Cache.StateSnapshots)-1]
			passed, err := r.DB.CheckNoResponsesForVersion(currentState.Version)
			if err != nil {
				err = fmt.Errorf("Error validating with db: %v", err)
			}
			if !passed {
				err = fmt.Errorf("Received a response for current version when we expected no response")
			}
			r.Service.Channels.Done <- true // send signal to close the channels and service down
			return err
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

func (r *Runner) DeltaClientReceivesOnlyTheResourceAndVersionForService(resource, version, service string) error {
	expectedResources := strings.Split(resource, ",")
	done := time.After(3 * time.Second)
	typeUrl := parser.ServiceToTypeURL(service)
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return fmt.Errorf("Err receiving responses, could not validate: %v", err)
		case <-done:
			passed, err := r.DB.DeltaCheckOnlyExpectedResources(expectedResources, version, typeUrl)
			if err != nil {
				return err
			}
			if !passed {
				return fmt.Errorf("Did not receive only the resource we wanted")
			}
			return nil
		}
	}
}

func (r *Runner) DeltaClientDoesNotReceiveResourceOfServiceAtVersion(resource, service, version string) error {
	done := time.After(15 * time.Second)
	typeUrl := parser.ServiceToTypeURL(service)
	log.Debug().
		Msg("Beginning 15 second grace period to check no updates are sent along the wire")
	for {
		select {
		case err := <-r.Service.Channels.Err:
			return fmt.Errorf("Err receiving responses, could not validate: %v", err)
		case <-done:
			passed, err := r.DB.DeltaCheckNoResource(resource, version, typeUrl)
			if err != nil {
				return err
			}
			if !passed {
				return fmt.Errorf("Expected no responses, but got responses matching resource and version")
			}
			log.Debug().
				Msg("No responses received within grace period")
			return nil
		}
	}
}

func (r *Runner) ResourceOfServiceIsUpdatedToVersion(resource, service, version string) error {
	typeUrl := parser.ServiceToTypeURL(service)
	c := pb.NewAdapterClient(r.Adapter.Conn)
	in := &pb.ResourceRequest{
		Node:         r.NodeID,
		TypeURL:      typeUrl,
		ResourceName: resource,
		Version:      version,
	}

	_, err := c.UpdateResource(context.Background(), in)
	if err != nil {
		msg := "Cannot uppdate resource using adapter"
		log.Error().
			Err(err).
			Msg(msg)
		return errors.New(msg)
	}
	log.Debug().
		Msgf("Updating resource %v to version %v", resource, version)
	return nil
}
func (r *Runner) ResourceIsRemovedFromTheService(resource, service string) error {
	typeUrl := parser.ServiceToTypeURL(service)
	c := pb.NewAdapterClient(r.Adapter.Conn)
	in := &pb.ResourceRequest{
		Node:         r.NodeID,
		TypeURL:      typeUrl,
		ResourceName: resource,
		Version:      "1",
	}

	_, err := c.RemoveResource(context.Background(), in)
	if err != nil {
		msg := "Cannnot remove resource using adapter"
		log.Error().Err(err).Msg(msg)
	}
	return nil
}

func (r *Runner) ResourceIsAddedToTheServiceAtVersion(resource, service, version string) error {
	typeUrl := parser.ServiceToTypeURL(service)
	c := pb.NewAdapterClient(r.Adapter.Conn)
	in := &pb.ResourceRequest{
		Node:         r.NodeID,
		TypeURL:      typeUrl,
		ResourceName: resource,
		Version:      version,
	}

	_, err := c.AddResource(context.Background(), in)
	if err != nil {
		msg := "Cannnot add resource using adapter"
		log.Error().Err(err).Msg(msg)
	}
	return nil
}

func (r *Runner) ClientUnsubscribesFromResourceForService(resource, service string) error {
	resources := strings.Split(resource, ",")
	typeUrl := parser.ServiceToTypeURL(service)
	request := &discovery.DeltaDiscoveryRequest{
		ResourceNamesUnsubscribe: resources,
		TypeUrl:                  typeUrl,
	}
	log.Debug().
		Msgf("Sending Unsubscribe Request: %v", request)
	r.Service.Channels.Delta_Req <- request
	return nil
}
