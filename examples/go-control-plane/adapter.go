package example

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"google.golang.org/grpc"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/golang/protobuf/ptypes"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
)

var (
	c cache.SnapshotCache
)

type Clusters  map[string]*cluster.Cluster

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

func (a *adapterServer) SetState (ctx context.Context, state *pb.Snapshot) (response *pb.SetStateResponse, err error) {

	clusters := make(map[string]*cluster.Cluster)

	// Parse Clusters
	for _, clstr := range state.Clusters.Items {
		seconds := time.Duration(clstr.ConnectTimeout["seconds"])
		clusters[clstr.Name] = &cluster.Cluster{
			Name: clstr.Name,
			ConnectTimeout: ptypes.DurationProto(seconds * time.Second),
		}
	}

	snapshot := cache.NewSnapshot(
		state.Version,
		// p.xdsCache.EndpointsContents(),
		[]types.Resource{}, // endpoints
		clusterContents(clusters), // clusters
		[]types.Resource{}, // routes
		[]types.Resource{}, // listeners
		[]types.Resource{}, // runtimes
		[]types.Resource{}, // secrets
	)
	if err = snapshot.Consistent(); err != nil {
		log.Printf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		os.Exit(1)
	}

	// // Add the snapshot to the cache
	if err := c.SetSnapshot(state.Node, snapshot); err != nil {
		log.Printf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
	newSnapshot, err := c.GetSnapshot(state.Node)
	fmt.Printf("new snapshot: \n%v\n\n", newSnapshot)
	response = &pb.SetStateResponse{
		Message: "Success",
	}
	return response, nil
}

func (a *adapterServer) ClearState(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	log.Printf("Clearing Cache")
	c.ClearSnapshot(req.Node)
	response := &pb.ClearResponse{
		Response: "All Clear",
	}
	return response, nil
}

func RunAdapter(port uint, cache cache.SnapshotCache) {
	c = cache
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
