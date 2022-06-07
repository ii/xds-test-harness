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
	connectTimeout = (90 * time.Second)
)

type Channels struct {
	Req  chan interface{} // will be a discoveryRequest or a deltaDiscoveryRequest
	Res  chan interface{} // will be a discoveryResponse or a deltadiscoveryResponse
	Err  chan error
	Done chan bool
}

type Context struct {
	context context.Context
	cancel  context.CancelFunc
}

type SotwStream interface {
	Send(*discovery.DiscoveryRequest) error
	Recv() (*discovery.DiscoveryResponse, error)
	CloseSend() error
}

type Sotw struct {
	Stream  SotwStream
	Context Context
}

type DeltaStream interface {
	Send(*discovery.DeltaDiscoveryRequest) error
	Recv() (*discovery.DeltaDiscoveryResponse, error)
	CloseSend() error
}

type Delta struct {
	Stream  DeltaStream
	Context Context
}

type XDSService struct {
	Name     string
	Channels *Channels
	Sotw     *Sotw
	Delta    *Delta
}

type serviceBuilder interface {
	openChannels()
	setSotwStream(conn *grpc.ClientConn) error
	setDeltaStream(conn *grpc.ClientConn) error
	getService(srv string) *XDSService
}

type LDSBuilder struct {
	Name     string
	Channels *Channels
	Sotw     *Sotw
	Delta    *Delta
}

func (b *LDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan interface{}, 2),
		Res:  make(chan interface{}, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *LDSBuilder) setSotwStream(conn *grpc.ClientConn) error {
	client := lds.NewListenerDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamListeners(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Sotw = &Sotw{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *LDSBuilder) setDeltaStream(conn *grpc.ClientConn) error {
	client := lds.NewListenerDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	delta, err := client.DeltaListeners(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Delta = &Delta{
		Stream: delta,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *LDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "LDS",
		Channels: b.Channels,
		Sotw:     b.Sotw,
		Delta:    b.Delta,
	}
}

type CDSBuilder struct {
	Name     string
	Channels *Channels
	Sotw     *Sotw
	Delta    *Delta
}

func (b *CDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan interface{}, 2),
		Res:  make(chan interface{}, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *CDSBuilder) setSotwStream(conn *grpc.ClientConn) error {
	client := cds.NewClusterDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamClusters(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Sotw = &Sotw{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *CDSBuilder) setDeltaStream(conn *grpc.ClientConn) error {
	client := cds.NewClusterDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.DeltaClusters(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Delta = &Delta{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *CDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "CDS",
		Channels: b.Channels,
		Sotw:     b.Sotw,
		Delta:    b.Delta,
	}
}

type RDSBuilder struct {
	Name     string
	Channels *Channels
	Sotw     *Sotw
	Delta    *Delta
}

func (b *RDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan interface{}, 2),
		Res:  make(chan interface{}, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *RDSBuilder) setSotwStream(conn *grpc.ClientConn) error {
	client := rds.NewRouteDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamRoutes(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Sotw = &Sotw{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *RDSBuilder) setDeltaStream(conn *grpc.ClientConn) error {
	client := rds.NewRouteDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.DeltaRoutes(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Delta = &Delta{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *RDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "RDS",
		Channels: b.Channels,
		Sotw:     b.Sotw,
		Delta:    b.Delta,
	}
}

type EDSBuilder struct {
	Name     string
	Channels *Channels
	Sotw     *Sotw
	Delta    *Delta
}

func (b *EDSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan interface{}, 2),
		Res:  make(chan interface{}, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *EDSBuilder) setSotwStream(conn *grpc.ClientConn) error {
	client := eds.NewEndpointDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamEndpoints(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Sotw = &Sotw{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *EDSBuilder) setDeltaStream(conn *grpc.ClientConn) error {
	client := eds.NewEndpointDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.DeltaEndpoints(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Delta = &Delta{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *EDSBuilder) getService(srv string) *XDSService {
	return &XDSService{
		Name:     "EDS",
		Channels: b.Channels,
		Sotw:     b.Sotw,
		Delta:    b.Delta,
	}
}

type ADSBuilder struct {
	Name     string
	Channels *Channels
	Sotw     *Sotw
	Delta    *Delta
}

func (b *ADSBuilder) openChannels() {
	b.Channels = &Channels{
		Req:  make(chan interface{}, 2),
		Res:  make(chan interface{}, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}

func (b *ADSBuilder) setSotwStream(conn *grpc.ClientConn) error {
	client := discovery.NewAggregatedDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.StreamAggregatedResources(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Sotw = &Sotw{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *ADSBuilder) setDeltaStream(conn *grpc.ClientConn) error {
	client := discovery.NewAggregatedDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	stream, err := client.DeltaAggregatedResources(ctx)
	if err != nil {
		defer cancel()
		return err
	}
	b.Delta = &Delta{
		Stream: stream,
		Context: Context{
			context: ctx,
			cancel:  cancel,
		},
	}
	return nil
}

func (b *ADSBuilder) getService(service string) *XDSService {
	return &XDSService{
		Name:     "ADS",
		Channels: b.Channels,
		Sotw:     b.Sotw,
		Delta:    b.Delta,
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
