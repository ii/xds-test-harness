package processor

import (
	"math"
	"os"
	"strconv"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/sirupsen/logrus"

	"github.com/zachmandeville/tester-prototype/examples/test-target/internal/resources"
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
			Listeners: make(map[string]resources.Listener),
			Clusters:  make(map[string]resources.Cluster),
			Routes:    make(map[string]resources.Route),
			Endpoints: make(map[string]resources.Endpoint),
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

func (p *Processor) UpdateSnapshot(cluster string) (snapshot cache.Snapshot, err error){

	p.xdsCache.AddCluster(cluster)

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
	p.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache

	if err := p.cache.SetSnapshot(p.nodeID, snapshot); err != nil {
		p.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
	return snapshot, err
}
