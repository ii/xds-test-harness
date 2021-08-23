package xdscache

import (
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/ptypes"

	pb "github.com/ii/xds-test-harness/api/adapter"
)

type XDSCache struct {
	Listeners map[string]Listener
	Routes    map[string]Route
	Clusters  map[string]*cluster.Cluster
	Endpoints map[string]Endpoint
}

func NewXDSCache() *XDSCache {
	return &XDSCache{
		Listeners: make(map[string]Listener),
		Clusters:  make(map[string]*cluster.Cluster),
		Routes:    make(map[string]Route),
		Endpoints: make(map[string]Endpoint),
	}
}

func (xds *XDSCache) ClusterContents() []types.Resource {
	var r []types.Resource
	for _, c := range xds.Clusters {
		r = append(r, c)
	}
	return r
}

func (xds *XDSCache) AddCluster(c *pb.Clusters_Cluster) {
	seconds := time.Duration(c.ConnectTimeout["seconds"])
	xds.Clusters[c.Name] = &cluster.Cluster{
		Name:           c.Name,
		ConnectTimeout: ptypes.DurationProto(seconds * time.Second),
	}
}
