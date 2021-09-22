package runner

import (
	"context"
	"errors"
	"sort"
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


func (r *Runner) ATargetSetupWithTheFollowingState(state *godog.DocString) error {
	snapshot, err := parser.YamlToSnapshot(r.NodeID, state.Content)
	if err != nil {
		msg := "Could not parse given state to adapter snapshot"
		log.Error().
			Stack().
			Err(err).
			Msg(msg)
		return errors.New(msg)
	}
	c := pb.NewAdapterClient(r.Adapter.Conn)
	_, err = c.SetState(context.Background(), snapshot)
	if err != nil {
		msg := "Cannot set target with given state"
		log.Error().
			Stack().
			Err(err).
			Msg(msg)
		return errors.New(msg)
	}
	r.Cache.StartState = snapshot
	return err
}

func (r *Runner) TheTargetIsUpdatedToTheFollowingState(state *godog.DocString) error {
	log.Debug().
		Msg("Updating target state")
	snapshot, err := parser.YamlToSnapshot(r.NodeID, state.Content)
	if err != nil {
		msg := "Could not parse given state to adapter snapshot"
		log.Error().
			Stack().
			Err(err).
			Msg(msg)
		return errors.New(msg)
	}
	c := pb.NewAdapterClient(r.Adapter.Conn)
	_, err = c.SetState(context.Background(), snapshot)
	if err != nil {
		msg := "Cannot set target with given state"
		log.Error().
			Stack().
			Err(err).
			Msg(msg)
		return errors.New(msg)
	}
	r.Cache.StartState = snapshot
	return err
}

func (r *Runner) ClientSubscribesToWildcardCDS() error {
	r.CDS.Req = make(chan *discovery.DiscoveryRequest, 1)
	r.CDS.Res = make(chan *discovery.DiscoveryResponse, 1)
	r.CDS.Err = make(chan error, 1)
	r.CDS.Done = make(chan bool, 1)
	r.CDS.Cache.InitResource = []string{}

	request := r.NewCDSRequest(r.CDS.Cache.InitResource)

	go r.CDSStream()
	go r.AckCDS(request)
	return nil
}

func (r *Runner) TheClientSubscribesToTheFollowingResources(resources *godog.DocString) error {
	resourceList, err := parser.ParseResourceList(resources.Content)
	if err != nil {
		log.Err(err).Msg("couldn't parse resource list")
	}
	r.CDS.Req = make(chan *discovery.DiscoveryRequest, 1)
	r.CDS.Res = make(chan *discovery.DiscoveryResponse, 1)
	r.CDS.Err = make(chan error, 1)
	r.CDS.Done = make(chan bool, 1)
	r.CDS.Cache.InitResource = resourceList

	request := r.NewCDSRequest(r.CDS.Cache.InitResource)

	go r.CDSStream()
	go r.AckCDS(request)
	return nil
}



func (r *Runner) ClientReceivesTheFollowingVersionAndClustersAlongWithNonce(resources *godog.DocString) error {
	expected, err := parser.YamlToSnapshot(r.NodeID, resources.Content)
	if err != nil {
		msg := "Couldn't parse test yaml. This is a problem with the test, not the target."
		log.Err(err).
			Msg(msg)
		return errors.New(msg)
	}

	for {
		select {
		case <-time.After(6 * time.Second):
			err := errors.New("Correct response not found after grace period of 6 seconds")
			log.Err(err).
				Msg("")
			return err
		default:
			if len(r.CDS.Cache.Responses) > 0 {
				for _, response := range r.CDS.Cache.Responses {
					actual, err := parser.ParseDiscoveryResponse(response)
					if err != nil {
						msg := "Could not parse Cached Response"
						log.Err(err).
							Msg(msg)
						return errors.New(msg)
					}
					if versionsMatch(expected, actual) && clustersMatch(expected, actual) {
						log.Debug().
							Msgf("Found Expected Response.\nexpected:%v\nactual: %v\n", expected, actual)
						return nil
					} else {
						err := errors.New("Expected Response does not match actual response.")
					    log.Err(err).
							Msgf("Expected: %v\nActual:%v", expected, actual)
						return err
					}
				}
			}
		}
	}
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

func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a target setup with the following state:$`, r.ATargetSetupWithTheFollowingState)
	ctx.Step(`^the Client subscribes to wildcard CDS$`, r.ClientSubscribesToWildcardCDS)
    ctx.Step(`^the Client subscribes to the following resources:$`, r.TheClientSubscribesToTheFollowingResources)
	ctx.Step(`^the Client receives the following version and clusters, along with a nonce:$`, r.ClientReceivesTheFollowingVersionAndClustersAlongWithNonce)
	ctx.Step(`^the Client sends an ACK to which the server does not respond$`, r.TheClientSendsAnACKToWhichTheServerDoesNotRespond)
    ctx.Step(`^the Target is updated to the following state:$`, r.TheTargetIsUpdatedToTheFollowingState)
}
