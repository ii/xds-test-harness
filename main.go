package main

import (
	"context"
	"encoding/json"
	//"flag"
	"fmt"
	"log"
	"google.golang.org/grpc"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
)

func main() {
	// Take a resource flag to practice sending different resources.
	// For simplicity, we will only request a single resource
	//resourceFlag := flag.String("resource", "foo", "a string")
	//flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The go-control-plane example server by default listens at :18000
	conn, err := grpc.Dial(":18000", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error connecting to management server: %v\n", err)
	}
	log.Printf("Connected to xDS Server. State: %v", conn.GetState())

	client := cluster_service.NewClusterDiscoveryServiceClient(conn)

	// Discovery Request following format of go-control-plane integration test.
	// we do not provide version_info to match the initial ACK diagram in
	// xDS protocol docs
	discoveryRequest := &envoy_service_discovery_v3.DiscoveryRequest{
		Node: &envoy_config_core_v3.Node{
			Id: "test-id",
		},
		TypeUrl:       resource.ClusterType,
		// Note that for CDS it is also possible to send a request w/o ResourceNames,
		// and it will return all clusters (wildcard request)
		ResourceNames: []string{"example_proxy_cluster"},
	}

	// Stream, send, and receive following integration test.
	sclient, err := client.StreamClusters(ctx)
	if err != nil {
		log.Fatalf("err setting up stream: %v", err.Error())
	}

	err = sclient.Send(discoveryRequest)
	if err != nil {
		log.Fatalf("err sending discoveryRequest: %v", err.Error())
	}

	awaitResponse := func() *envoy_service_discovery_v3.DiscoveryResponse {
		doneCh := make(chan *envoy_service_discovery_v3.DiscoveryResponse)
		go func() {

			r, err := sclient.Recv()
			if err != nil {
				fmt.Printf("errrrr: %v", err.Error())
			}
			doneCh <- r
		}()
		return <-doneCh
	}

	discoveryResponse := awaitResponse()
	respJSON, err := json.MarshalIndent(discoveryResponse, "", "   ")
	if err != nil {
		log.Fatalf("Error marshalling discoveryResponse: %v", err.Error())
	}
	log.Printf("response: %v\n", string(respJSON))
}
