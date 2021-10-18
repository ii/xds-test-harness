package runner

import (
	// "fmt"
	// "context"
	// core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	// cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
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


type serviceBuilder interface {
	openChannels()
	setStreamFn()
	setInitResources([]string)
	getService() *xDSService

}

type xDSService struct {
	Name string
	TypeURL string
	Channels *Channels
	Cache *ServiceCache
	Startfn func(*grpc.ClientConn) (interface{})
}

type LDSBuilder struct {
	Name string
	TypeURL string
	Channels *Channels
	Cache *ServiceCache
	Startfn func(*grpc.ClientConn) (interface{})
}

func (b *LDSBuilder) openChannels () {
	b.Channels = &Channels{
		Req:  make(chan *discovery.DiscoveryRequest, 2),
		Res:  make(chan *discovery.DiscoveryResponse, 2),
		Err:  make(chan error, 2),
		Done: make(chan bool),
	}
}


func (b *LDSBuilder) setStreamFn() {
	b.Startfn = func(conn *grpc.ClientConn) interface{}{
	  client := lds.NewListenerDiscoveryServiceClient(conn)
		return client.StreamListeners
	}
}

func (b *LDSBuilder) setInitResources (res []string) {
	b.Cache.InitResource = res
}

func (b *LDSBuilder) getService () *xDSService {
	return &xDSService{
		Name:     "LDS",
		TypeURL:  TypeUrlLDS,
		Channels: b.Channels,
		Cache: b.Cache,
		Startfn: b.Startfn,
	}
}

// type CDSBuilder struct {
// 	Name string
// 	TypeURL string
// 	Channels *Channels
// 	Cache *Cache
// }

// func (b *CDSBuilder) openChannels () {
// 	b.Channels = &Channels{
// 		Req:  make(chan *discovery.DiscoveryRequest),
// 		Res:  make(chan *discovery.DiscoveryResponse),
// 		Err:  make(chan error),
// 		Done: make(chan bool),
// 	}
// }

// func (b *CDSBuilder) startStream (ctx context.Context, conn *grpc.ClientConn)  (interface{}, error) {
// 	client := cds.NewClusterDiscoveryServiceClient(conn)
// 	stream, err := client.StreamClusters(ctx)
// 	return stream, err
// }

// func (b *CDSBuilder) getService () *xDSService {
// 	return &xDSService{
// 		Name:     "CDS",
// 		TypeURL:  TypeUrlCDS,
// 		Channels: &Channels{},
// 		Cache:    &Cache{},
// 	}
// }

func getBuilder(builderType string) serviceBuilder {
	if builderType == "LDS" {
		return &LDSBuilder{}
	}
	// if builderType == "CDS" {
	// 	return &CDSBuilder{}
	// }
	return nil
}
