package xdscache

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/zachmandeville/tester-prototype/test-target/internal/resources"
)

type XDSCache struct {
	Listeners map[string]resources.Listener
	Routes    map[string]resources.Route
	Clusters  map[string]resources.Cluster
	Endpoints map[string]resources.Endpoint
}

func (xds *XDSCache) ClusterContents() []types.Resource {
	var r []types.Resource

	for _, c := range xds.Clusters {
		r = append(r, resources.MakeCluster(c.Name))
	}
	return r
}

func (xds *XDSCache) AddCluster(name string) {
	xds.Clusters[name] = resources.Cluster{
		Name: name,
	}
}
