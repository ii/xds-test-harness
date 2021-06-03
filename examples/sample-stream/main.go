package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	// "protobuf/jsonb"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
)

const target = ":18000"

var (
	done = make(chan bool)
)

type Runner struct {
	ResC chan *envoy_service_discovery_v3.DiscoveryResponse
	Conn *grpc.ClientConn
}

func (r *Runner) connectToTarget() error {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return err
	}
	r.Conn = conn
	fmt.Printf("Connected to target at %v\n", target)
	return nil
}

func (r *Runner) startStream() {
	dreq := &envoy_service_discovery_v3.DiscoveryRequest{
		VersionInfo: "",
		Node: &envoy_config_core_v3.Node{
			Id: "test-id",
		},
		TypeUrl: "type.googleapis.com/envoy.config.cluster.v3.Cluster",
	}
	c := cluster_service.NewClusterDiscoveryServiceClient(r.Conn)
	stream, err := c.StreamClusters(context.Background())
	if err != nil {
		fmt.Printf("error streaming clusters: %v", err)
	}
	stream.Send(dreq)
	go func() {
		time.Sleep(5 * time.Second)
		stream.CloseSend()
	}()
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("what error: %v", err)
			done <- true
			break
		}
		if err != nil {
			fmt.Println("what error 2!")
			break
		}
		resources := res.GetResources()
		var m cluster.Cluster
		err = ptypes.UnmarshalAny(resources[0], &m)
		if err != nil {
			fmt.Printf("anypb error: %v\n", err)
		}
		mj, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("error marshaling to json: %v", err)
		}
		fmt.Printf("anypb resources: %v\n", string(mj))
		r.ResC <- res
	}
}

func main() {
	r := Runner{}
	r.ResC = make(chan *envoy_service_discovery_v3.DiscoveryResponse)
	fmt.Printf("gotta channel: %v\n'", r.ResC)
	if err := r.connectToTarget(); err != nil {
		fmt.Printf("ERRROR!: %v", err)
	}
	go r.startStream()
	for {
		select {
		case res := <-r.ResC:
			fmt.Printf("res from chan %v\n", res)
		case <-done:
			println("Done")
			close(r.ResC)
			return
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}
