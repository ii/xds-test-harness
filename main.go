package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"io"
	"google.golang.org/grpc"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
)

func createRequest (version string, nonce string) *envoy_service_discovery_v3.DiscoveryRequest {
	return &envoy_service_discovery_v3.DiscoveryRequest{
		VersionInfo: version,
		Node: &envoy_config_core_v3.Node{
			Id: "test-id",
		},
		TypeUrl: resource.ClusterType,
		// Note that for CDS it is also possible to send a request w/o ResourceNames,
		// and it will return all clusters (wildcard request)
		// ResourceNames: []string{},
		ResponseNonce: nonce,
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The go-control-plane example server by default listens at :18000
	conn, err := grpc.Dial(":18000", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error connecting to management server: %v\n", err)
	}

	client := cluster_service.NewClusterDiscoveryServiceClient(conn)

	// Stream, send, and receive following integration test.
	stream, err := client.StreamClusters(ctx)
	if err != nil {
		log.Fatalf("err setting up stream: %v", err.Error())
	}

	waitc := make(chan *envoy_service_discovery_v3.DiscoveryResponse)

	// Start the receiving stream, coming from the xDS server. All responses sent
	// to the waitc channel
	go func () {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive Discovery Response: %v", err.Error())
			}

			// pretty-print to stdout
			responseJSON, err := json.MarshalIndent(in, "", " ")
			if err != nil {
				log.Fatalf("Error marshalling response: %v", err)
			}
			log.Printf("Got Response: %v", string(responseJSON))

			waitc <- in
		}
	}()

	dreq := createRequest("","")

	// Pretty print this too.
	requestJSON, err:= json.MarshalIndent(dreq, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling discovery request: %v", err.Error())
	}

	fmt.Printf("sending DiscoveryRequest:\n%v\n ", string(requestJSON))
	if err = stream.Send(dreq); err != nil {
		log.Fatalf("err sending discoveryRequest: %v", err.Error())
	}

	// set up our last known version, which will be the empty string we sent in our initial discovery request.
	last_version := ""

	// endless loop until signal interrupt.
	// We take the latest discovery response and, if there's new version_info, send a new
	// discovery request confirming we've received response successfully.
	for {
      dres := <-waitc
		if dres.VersionInfo != last_version {
			dreq = createRequest(dres.VersionInfo, dres.Nonce)
			requestJSON, err:= json.MarshalIndent(dreq, "", "  ")
			if err != nil {
				log.Fatalf("error marshalling discovery request: %v", err.Error())
			}

			fmt.Printf("sending DiscoveryRequest:\n%v\n ", string(requestJSON))
			if err = stream.Send(dreq); err != nil {
				log.Fatalf("err sending discoveryRequest: %v", err.Error())
			}
			// this is a sanity check. Since we are communicating with CDS, we could expect that if new clusters are added,
			// then we should see a new version and a new number of resources from previous.
	        log.Printf("\nLast Version: %v, \nNew Version: %v,\nResources: %v\n", last_version, dres.VersionInfo, len(dres.GetResources()))
			last_version = dres.VersionInfo
		}
	}
	//TODO end this gracefully
}
