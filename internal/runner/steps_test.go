package runner

import (
	// listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	// route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3" discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	// parser "github.com/ii/xds-test-harness/internal/parser"
	// "google.golang.org/protobuf/proto"
	// "google.golang.org/protobuf/types/known/anypb"
	"testing"
)

func TestClientReceivesCorrectResourceVersionService(t *testing.T) {
	// yah := "kakapo,tui,kaka"
	// // nah := "gecko,salamander,tui"
	// srv := "LDS"

	// expected := []string{"kakapo", "tui", "kaka"}
	// listeners := []*anypb.Any{}
	// for _, name := range expected {
	// 	dst := &anypb.Any{}
	// 	src := &listener.Listener{Name: name}
	// 	opts := proto.MarshalOptions{}
	// 	err := anypb.MarshalFrom(dst, src, opts)
	// 	if err != nil {
	// 		t.Errorf("Error marshalling listener to anypb.any: %v", err)
	// 	}
	// 	listeners = append(listeners, dst)
	// }
	// ldsResponse := &discovery.DiscoveryResponse{
	// 	VersionInfo: "1",
	// 	Resources:   listeners,
	// 	TypeUrl:     parser.TypeUrlLDS,
	// 	Nonce:       "1",
	// }
	// runner := FreshRunner()
	// runner.Service = &XDSService{
	// 	Name:     srv,
	// 	Channels: &Channels{},
	// }

	// err := runner.ClientReceivesResourcesAndVersionForService(yah, "1", "LDS")
	// if err != nil {
	// 	t.Errorf("Could not find response in cache, though it should be there. err :%v", err)
	// }
	// // for RDS or EDS, resources can come in multiple responses.

	// rdsResponses := []*discovery.DiscoveryResponse{}
	// for _, name := range expected {
	// 	routes := []*anypb.Any{}
	// 	dst := &anypb.Any{}
	// 	src := &route.RouteConfiguration{Name: name}
	// 	opts := proto.MarshalOptions{}
	// 	err := anypb.MarshalFrom(dst, src, opts)
	// 	if err != nil {
	// 		t.Errorf("Error marshalling listener to anypb.any: %v", err)
	// 	}
	// 	routes = append(routes, dst)
	// 	rdsResponse := &discovery.DiscoveryResponse{
	// 		VersionInfo: "1",
	// 		Resources:   routes,
	// 		TypeUrl:     parser.TypeUrlRDS,
	// 		Nonce:       "1",
	// 	}
	// 	rdsResponses = append(rdsResponses, rdsResponse)
	// }
	// runner = FreshRunner()
	// runner.Service = &XDSService{
	// 	Name:     "RDS",
	// 	Channels: &Channels{},
	// 	Cache:    &ServiceCache{},
	// 	Stream:   nil,
	// 	Context:  Context{},
	// }
	// runner.Service.Cache.Responses = rdsResponses
	// err = runner.ClientReceivesResourcesAndVersionForService(yah, "1", "RDS")
	// if err != nil {
	// 	t.Errorf("Could not find response in cache for RDS, thought it was added. %v", err)
	// }
}

func TestClientReceivesOnlyTheCorrectResource(t *testing.T) {
	// For RDS or EDS, when we subscribe to a single resource only a single resource should be in response.
	// yah := "kakapo"
	// routes := []*anypb.Any{}
	// dst := &anypb.Any{}
	// src := &route.RouteConfiguration{Name: yah}
	// opts := proto.MarshalOptions{}
	// err := anypb.MarshalFrom(dst, src, opts)
	// if err != nil {
	// 	t.Errorf("Error marshalling listener to anypb.any: %v", err)
	// }
	// routes = append(routes, dst)
	// rdsResponse := &discovery.DiscoveryResponse{
	// 	VersionInfo: "1",
	// 	Resources:   routes,
	// 	TypeUrl:     parser.TypeUrlRDS,
	// }
	// runner := FreshRunner()

	// runner.Service = &XDSService{
	// 	Name:     "RDS",
	// 	Channels: &Channels{},
	// 	Cache:    &ServiceCache{},
	// 	Stream:   nil,
	// 	Context:  Context{},
	// }
	// runner.Service.Cache.Responses = append(runner.Service.Cache.Responses, rdsResponse)
	// err = runner.ClientReceivesOnlyTheCorrectResourceAndVersion("kakapo", "1")
	// if err != nil {
	// 	t.Errorf("Client received more than we expected. %v", runner.Service.Cache.Responses)
	// }

}
