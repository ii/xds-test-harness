package runner

import (
	"testing"

	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestFreshRunner(t *testing.T) {
	fresh := FreshRunner()
	// fresh runner should not have any adapter set
	if fresh.Adapter.Port != "" || fresh.Target.Conn != nil || fresh.Aggregated != false {
		t.Errorf("Fresh runner has values that should be empty. Runner %v", fresh)
	}

	fresh.Aggregated = true
	fresh.Adapter.Port = ":18000"
	fresh.NodeID = "tui"

	redo := FreshRunner(fresh)
	if redo.Aggregated != true || redo.Adapter.Port != ":18000" || redo.NodeID != "tui" {
		t.Errorf("Fresh Runner did not use the values passed into it. Example, node id: %v", redo.NodeID)
	}
}

// Ack should be building a cache of responses
// and requests, and sending back a new request per response
// received on the channel.
func TestAck(t *testing.T) {

	r := FreshRunner()
	r.NodeID = "testing"

	// create the channels with arbitrary service,
	// and arbitrary starting resources
	service := "LDS"
	typeURL := "type.googleapis.com/envoy.config.listener.v3.Listener"
	resources := []string{"tui", "kaka", "kakapo"}
	builder := getBuilder(service)
	builder.openChannels()
	builder.setInitResources(resources)
	r.Service = builder.getService(service)

	request := newRequest(resources, typeURL, r.NodeID)
	// start the loop with basic request
	r.SubscribeRequest = request
	go r.Ack(r.Service)

	listeners := []*anypb.Any{}
	for _, name := range resources {
		dst := &anypb.Any{}
		src := &listener.Listener{Name: name}
		opts := proto.MarshalOptions{}
		err := anypb.MarshalFrom(dst, src, opts)
		if err != nil {
			t.Errorf("Error marshalling listener to anypb.any: %v", err)
		}
		listeners = append(listeners, dst)
	}

	response := &discovery.DiscoveryResponse{
		VersionInfo: "1",
		Resources:   listeners,
		TypeUrl:     typeURL,
		// Nonce:        "1",
	}

	// mock a response received
	r.Service.Channels.Res <- response

	// flush out the request channel
	// (in practice, this is done by our Stream fn)
	req := <-r.Service.Channels.Req
	log.Debug().Msgf("request %v", req)

	// pass a second response to make sure our cache
	// and channels can handle multiple responses
	secondResponse := &discovery.DiscoveryResponse{
		VersionInfo: "2",
		Resources:   listeners,
		TypeUrl:     typeURL,
	}

	r.Service.Channels.Res <- secondResponse
	req = <-r.Service.Channels.Req
	// send a done request which should close Ack
	// and stop its running
	r.Service.Channels.Done <- true

	// At this point, we should have three requests and two responses
	// in our cache (the initial request, the responses above, and an ack
	// per response)
	responses := r.Service.Cache.Responses
	requests := r.Service.Cache.Requests
	if len(responses) != 2 {
		t.Errorf("Ack did not cache the expecteda mount of responses. expected 2, actual %v", len(responses))
	}
	if len(requests) != 3 {
		t.Errorf("Ack did not cache the expecteda mount of responses. expected 3, actual %v", len(requests))
	}
	// the cache should have the responses in order too
	if responses[0].VersionInfo != "1" || responses[1].VersionInfo != "2" {
		t.Error("Responses were not received and/or cached in order. Cannot trust sequence happened correctly")
	}
}
