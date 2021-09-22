package example

import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	"net"
	"os"
	"time"
	"google.golang.org/grpc"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/golang/protobuf/ptypes"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

var (
	xdsCache cache.SnapshotCache
)

type Clusters  map[string]*cluster.Cluster
type Listeners map[string]*listener.Listener

type adapterServer struct {
	pb.UnimplementedAdapterServer
}


func clusterContents(clusters Clusters) []types.Resource {
	var r []types.Resource
	for _, c := range clusters {
		r = append(r, c)
	}
	return r
}

func listenerContents(listeners Listeners) []types.Resource {
	var r []types.Resource
	for _, l := range listeners {
		r = append(r,l)
	}
	return r
}

func (a *adapterServer) SetState (ctx context.Context, state *pb.Snapshot) (response *pb.SetStateResponse, err error) {


	// Parse Clusters
	clusters := make(map[string]*cluster.Cluster)
	for _, c := range state.Clusters.Items {
		seconds := time.Duration(c.ConnectTimeout["seconds"])
		clusters[c.Name] = &cluster.Cluster{
			Name: c.Name,
			ConnectTimeout: ptypes.DurationProto(seconds * time.Second),
		}
	}

	// Parse Listeners
	listeners := make(map[string]*listener.Listener)
	if state.Listeners != nil {

		for _, l := range state.Listeners.Items {
			listeners[l.Name] = &listener.Listener{
				Name:                             l.Name,
				Address:                          &core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Address: l.Address,
						},
					},
				},
			}
		}
	}

	snapshot := cache.NewSnapshot(
		state.Version,
		// p.xdsCache.EndpointsContents(),
		[]types.Resource{}, // endpoints
		clusterContents(clusters), // clusters
		[]types.Resource{}, // routes
		listenerContents(listeners),
		[]types.Resource{}, // runtimes
		[]types.Resource{}, // secrets
	)
	if err = snapshot.Consistent(); err != nil {
		log.Printf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		os.Exit(1)
	}

	// // Add the snapshot to the cache
	if err := xdsCache.SetSnapshot(state.Node, snapshot); err != nil {
		log.Printf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
	newSnapshot, err := xdsCache.GetSnapshot(state.Node)
	prettySnap, _ := json.Marshal(newSnapshot)
	fmt.Printf("new snapshot: \n%v\n\n", string(prettySnap))
	response = &pb.SetStateResponse{
		Message: "Success",
	}
	return response, nil
}

func (a *adapterServer) ClearState(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	log.Printf("Clearing Cache")
	xdsCache.ClearSnapshot(req.Node)
	response := &pb.ClearResponse{
		Response: "All Clear",
	}
	return response, nil
}

func RunAdapter(port uint, cache cache.SnapshotCache) {
	xdsCache = cache
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterAdapterServer(s, &adapterServer{})
	log.Printf("Testsuite Adapter listening on %v\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Adapter failed to serve: %v", err)
	}
}
