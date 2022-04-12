package runner

import (
	"context"
	"time"

	cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	eds "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	lds "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	rds "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"google.golang.org/grpc"
)

type Channels struct {
	Req       chan *discovery.DiscoveryRequest
	Res       chan *discovery.DiscoveryResponse
	Delta_Req chan *discovery.DeltaDiscoveryRequest
	Delta_Res chan *discovery.DeltaDiscoveryResponse
	Err       chan error
	Done      chan bool
}

type ServiceCache struct {
	InitResource    []string
	Requests        []*discovery.DiscoveryRequest
	Responses       []*discovery.DiscoveryResponse
	Delta_Requests  []*discovery.DeltaDiscoveryRequest
	Delta_Responses []*discovery.DeltaDiscoveryResponse
}

type Context struct {
	context context.Context
	cancel  context.CancelFunc
}

type Stream interface {
	Send(*discovery.DiscoveryRequest) error
	Recv() (*discovery.DiscoveryResponse, error)
	CloseSend() error
}

type DeltaStream interface {
	Send(*discovery.DeltaDiscoveryRequest) error
	Recv() (*discovery.DeltaDiscoveryResponse, error)
	CloseSend() error
}

type XDSService struct {
	Name     string
	Channels *Channels
	Cache    *ServiceCache
	Stream   Stream
	Delta    DeltaStream
	Context  Context
}

type serviceBuilder interface {
	openChannels()
	setStreams(conn *grpc.ClientConn) error
	setInitResources([]string)
	getService(srv string) *XDSService
}

type LDSBuilder struct {
	Name     string
	Channels *Channels
	Cache    *ServiceCache
	Stream   Stream
	Delta    DeltaStream
	Context  Context
}

func (b *LDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Delta_Req:  make(chan *discovery.DeltaDiscoveryRequest, 2),
		Delta_Res:  make(chan *discovery.DeltaDiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *LDSBuilder) setStreams(conn *grpc.ClientConn) error {
	client := lds.NewListenerDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := client.StreamListeners(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	delta, err := client.DeltaListeners(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	b.Delta = delta
	return nil
}

func (b *LDSBuilder) setInitResources(res []string) {
	b.Cache = &ServiceCache{}
	b.Cache.InitResource = res
}

func (b *LDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "LDS",
		Channels: b.Channels,
		Cache:    b.Cache,
		Stream:   b.Stream,
		Delta:    b.Delta,
	}
}

type CDSBuilder struct {
	Name     string
	Channels *Channels
	Cache    *ServiceCache
	Stream   Stream
	Delta DeltaStream
	Context  Context
}

func (b *CDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Delta_Req:  make(chan *discovery.DeltaDiscoveryRequest, 2),
		Delta_Res:  make(chan *discovery.DeltaDiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *CDSBuilder) setStreams(conn *grpc.ClientConn) error {
	client := cds.NewClusterDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := client.StreamClusters(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	delta, err := client.DeltaClusters(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	b.Delta = delta
	return nil
}

func (b *CDSBuilder) setInitResources(res []string) {
	b.Cache = &ServiceCache{}
	b.Cache.InitResource = res
}

func (b *CDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "CDS",
		Channels: b.Channels,
		Cache:    b.Cache,
		Stream:   b.Stream,
		Delta: b.Delta,
	}
}

type RDSBuilder struct {
	Name     string
	Channels *Channels
	Cache    *ServiceCache
	Stream   Stream
	Delta DeltaStream
	Context  Context
}

func (b *RDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Delta_Req:  make(chan *discovery.DeltaDiscoveryRequest, 2),
		Delta_Res:  make(chan *discovery.DeltaDiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *RDSBuilder) setStreams(conn *grpc.ClientConn) error {
	client := rds.NewRouteDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := client.StreamRoutes(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	delta, err := client.DeltaRoutes(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	b.Delta = delta
	return nil
}

func (b *RDSBuilder) setInitResources(res []string) {
	b.Cache = &ServiceCache{}
	b.Cache.InitResource = res
}

func (b *RDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "RDS",
		Channels: b.Channels,
		Cache:    b.Cache,
		Stream:   b.Stream,
		Delta: b.Delta,
	}
}

type EDSBuilder struct {
	Name     string
	Channels *Channels
	Cache    *ServiceCache
	Stream   Stream
	Delta DeltaStream
	Context  Context
}

func (b *EDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Delta_Req:  make(chan *discovery.DeltaDiscoveryRequest, 2),
		Delta_Res:  make(chan *discovery.DeltaDiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *EDSBuilder) setStreams(conn *grpc.ClientConn) error {
	client := eds.NewEndpointDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := client.StreamEndpoints(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	delta, err := client.DeltaEndpoints(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	b.Delta = delta
	return nil
}

func (b *EDSBuilder) setInitResources(res []string) {
	b.Cache = &ServiceCache{}
	b.Cache.InitResource = res
}

func (b *EDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "EDS",
		Channels: b.Channels,
		Cache:    b.Cache,
		Stream:   b.Stream,
		Delta: b.Delta,
	}
}

type ADSBuilder struct {
	Name     string
	Channels *Channels
	Cache    *ServiceCache
	Stream   Stream
	Delta DeltaStream
	Context  Context
}

func (b *ADSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Delta_Req:  make(chan *discovery.DeltaDiscoveryRequest, 2),
		Delta_Res:  make(chan *discovery.DeltaDiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *ADSBuilder) setStreams(conn *grpc.ClientConn) error {
	client := discovery.NewAggregatedDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := client.StreamAggregatedResources(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	delta, err := client.DeltaAggregatedResources(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	b.Delta = delta
	return nil
}

func (b *ADSBuilder) setInitResources(res []string) {
	b.Cache = &ServiceCache{}
	b.Cache.InitResource = res
}

func (b *ADSBuilder) getService(service string) *XDSService {
	return &XDSService{
		Name:     "ADS",
		Channels: b.Channels,
		Cache:    b.Cache,
		Stream:   b.Stream,
		Delta: b.Delta,
	}
}

func getBuilder(builderType string) serviceBuilder {
	switch builderType {
	case "LDS":
		return &LDSBuilder{}
	case "CDS":
		return &CDSBuilder{}
	case "RDS":
		return &RDSBuilder{}
	case "EDS":
		return &EDSBuilder{}
	case "ADS":
		return &ADSBuilder{}
	default:
		return nil
	}
}
