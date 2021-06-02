package main

import (
	"context"
	"flag"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	logrus "github.com/sirupsen/logrus"

	"github.com/zachmandeville/tester-prototype/examples/test-target/internal/adapter"
	"github.com/zachmandeville/tester-prototype/examples/test-target/internal/processor"
	"github.com/zachmandeville/tester-prototype/examples/test-target/internal/server"
)

var (
	nodeID      string
	l           logrus.FieldLogger
	adapterPort string
	port        uint
	proc        *processor.Processor
)

func init() {
	l = logrus.New()
	logrus.SetLevel(logrus.DebugLevel)
	flag.UintVar(&port, "port", 18000, "xDS management server port")
	flag.StringVar(&nodeID, "nodeID", "test-id", "NodeID")
	flag.StringVar(&adapterPort, "adapterPort", ":17000", "port of test suite adapter")
}

func main() {
	flag.Parse()
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)
	proc = processor.NewProcessor(
		cache, nodeID, logrus.WithField("context", "processor"))
	go func() {
		// Run the xDS server
		ctx := context.Background()
		srv := serverv3.NewServer(ctx, cache, nil)
		server.RunServer(ctx, srv, port)
	}()

	adapter.RunServer(proc, adapterPort)
}
