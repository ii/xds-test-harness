package processor

import (
	"math"
	"os"
	"strconv"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/sirupsen/logrus"

	pb "github.com/zachmandeville/tester-prototype/api/adapter"
	"github.com/zachmandeville/tester-prototype/examples/test-target/internal/xdscache"
)

type Processor struct {
	cache           cache.SnapshotCache
	nodeID          string
	snapshotVersion int64
	logrus.FieldLogger
	xdsCache xdscache.XDSCache
}

func NewProcessor(cache cache.SnapshotCache, nodeID string, log logrus.FieldLogger) *Processor {
	return &Processor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: 1,
		FieldLogger:     log,
		xdsCache: xdscache.XDSCache{
			Listeners: make(map[string]xdscache.Listener),
			Clusters:  make(map[string]*cluster.Cluster),
			Routes:    make(map[string]xdscache.Route),
			Endpoints: make(map[string]xdscache.Endpoint),
		},
	}
}

func (p *Processor) newSnapshotVersion() string {
	//reset if it number gets too high, and make sure our first snapshot is version 1
	if p.snapshotVersion == math.MaxInt64 || p.snapshotVersion == 1 {
		p.snapshotVersion = 0
	}
	p.snapshotVersion++
	return strconv.FormatInt(p.snapshotVersion, 10)
}

func (p *Processor) UpdateSnapshot(state *pb.Snapshot) (snapshot cache.Snapshot, err error) {

	// Parse Clusters
	for _, c := range state.Clusters.Items {
		p.xdsCache.AddCluster(c)
	}

	snapshot = cache.NewSnapshot(
		p.newSnapshotVersion(),
		// p.xdsCache.EndpointsContents(),
		[]types.Resource{}, // endpoints
		p.xdsCache.ClusterContents(),
		[]types.Resource{}, // routes
		// p.xdsCache.RouteContents(),     // routes
		[]types.Resource{}, // listeners
		// p.xdsCache.ListenerContents(),  // listeners
		[]types.Resource{}, // runtimes
		[]types.Resource{}, // secrets
	)

	if err = snapshot.Consistent(); err != nil {
		p.Errorf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		return
	}
	p.Debugf(" will serve snapshot:\n%+v\n\n", snapshot)

	// Add the snapshot to the cache

	if err := p.cache.SetSnapshot(state.Node, snapshot); err != nil {
		p.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
	return snapshot, err
}
