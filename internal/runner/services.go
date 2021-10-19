package runner

import (
	// "fmt"
	"context"
	"time"
	// core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	lds "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
)

const (
	TypeUrlLDS= "type.googleapis.com/envoy.config.listener.v3.Listener"
	TypeUrlCDS = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	// TYPEURL_RDS = ""
)

type Channels struct {
	Req   chan *discovery.DiscoveryRequest
	Res   chan *discovery.DiscoveryResponse
	Err   chan error
	Done  chan bool
}

type ServiceCache struct {
	InitResource []string
	Requests  []*discovery.DiscoveryRequest
	Responses []*discovery.DiscoveryResponse
}

type Context struct {
	context context.Context
	cancel context.CancelFunc
}

type Stream interface {
	Send(*discovery.DiscoveryRequest) error
	Recv() (*discovery.DiscoveryResponse, error)
	CloseSend() error
}


type XDSService struct {
	Name string
	TypeURL string
	Channels *Channels
	Cache *ServiceCache
	Stream Stream
	Context Context
}

type serviceBuilder interface {
	openChannels()
	setStream(conn *grpc.ClientConn) error
	setInitResources([]string)
	getService() *XDSService

}

type LDSBuilder struct {
	Name string
	TypeURL string
	Channels *Channels
	Cache *ServiceCache
	Stream Stream
	Context Context
}

func (b *LDSBuilder) openChannels () {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}


func (b *LDSBuilder) setStream(conn *grpc.ClientConn) error {
	client := lds.NewListenerDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := client.StreamListeners(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	return nil
}

func (b *LDSBuilder) setInitResources (res []string) {
	b.Cache = &ServiceCache{}
	b.Cache.InitResource = res
}

func (b *LDSBuilder) getService () *XDSService {
	return &XDSService{
		Name:     "LDS",
		TypeURL:  TypeUrlLDS,
		Channels: b.Channels,
		Cache: b.Cache,
		Stream: b.Stream,
	}
}

type CDSBuilder struct {
	Name string
	TypeURL string
	Channels *Channels
	Cache *ServiceCache
	Stream Stream
	Context Context
}

func (b *CDSBuilder) openChannels () {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}


func (b *CDSBuilder) setStream(conn *grpc.ClientConn) error {
	client := cds.NewClusterDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := client.StreamClusters(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	return nil
}

func (b *CDSBuilder) setInitResources (res []string) {
	b.Cache = &ServiceCache{}
	b.Cache.InitResource = res
}

func (b *CDSBuilder) getService () *XDSService {
	return &XDSService{
		Name:     "CDS",
		TypeURL:  TypeUrlCDS,
		Channels: b.Channels,
		Cache: b.Cache,
		Stream: b.Stream,
	}
}

func getBuilder(builderType string) serviceBuilder {
	if builderType == "LDS" {
		return &LDSBuilder{}
	}
	if builderType == "CDS" {
		return &CDSBuilder{}
	}
	return nil
}
