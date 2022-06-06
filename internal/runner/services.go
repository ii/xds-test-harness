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

const (
	connectTimeout = (60 * time.Second)
)

type Channels struct {
	Req  chan *discovery.DiscoveryRequest
	Res  chan *discovery.DiscoveryResponse
	Err  chan error
	Done chan bool
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

type XDSService struct {
	Name     string
	Channels *Channels
	Stream   Stream
	Context  Context
}

type serviceBuilder interface {
	openChannels()
	setStream(conn *grpc.ClientConn) error
	getService(srv string) *XDSService
}

type LDSBuilder struct {
	Name     string
	Channels *Channels
	Stream   Stream
	Context  Context
}

func (b *LDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *LDSBuilder) setStream(conn *grpc.ClientConn) error {
	client := lds.NewListenerDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
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

func (b *LDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "LDS",
		Channels: b.Channels,
		Stream:   b.Stream,
	}
}

type CDSBuilder struct {
	Name     string
	Channels *Channels
	Stream   Stream
	Context  Context
}

func (b *CDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *CDSBuilder) setStream(conn *grpc.ClientConn) error {
	client := cds.NewClusterDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
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

func (b *CDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "CDS",
		Channels: b.Channels,
		Stream:   b.Stream,
	}
}

type RDSBuilder struct {
	Name     string
	Channels *Channels
	Stream   Stream
	Context  Context
}

func (b *RDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *RDSBuilder) setStream(conn *grpc.ClientConn) error {
	client := rds.NewRouteDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamRoutes(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	return nil
}

func (b *RDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "RDS",
		Channels: b.Channels,
		Stream:   b.Stream,
	}
}

type EDSBuilder struct {
	Name     string
	Channels *Channels
	Stream   Stream
	Context  Context
}

func (b *EDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *EDSBuilder) setStream(conn *grpc.ClientConn) error {
	client := eds.NewEndpointDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamEndpoints(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	return nil
}

func (b *EDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "EDS",
		Channels: b.Channels,
		Stream:   b.Stream,
	}
}

type ADSBuilder struct {
	Name     string
	Channels *Channels
	Stream   Stream
	Context  Context
}

func (b *ADSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *ADSBuilder) setStream(conn *grpc.ClientConn) error {
	client := discovery.NewAggregatedDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamAggregatedResources(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Context.context = ctx
	b.Context.cancel = cancel
	b.Stream = stream
	return nil
}

func (b *ADSBuilder) getService(service string) *XDSService {
	return &XDSService{
		Name:     "ADS",
		Channels: b.Channels,
		Stream:   b.Stream,
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
