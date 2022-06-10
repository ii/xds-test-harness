package example

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	runtime "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	TypeUrlLDS = "type.googleapis.com/envoy.config.listener.v3.Listener"
	TypeUrlCDS = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	TypeUrlRDS = "type.googleapis.com/envoy.config.route.v3.RouteConfiguration"
	TypeUrlEDS = "type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment"
)

var (
	xdsCache  cache.SnapshotCache
	localhost = "127.0.0.1"
)

type Clusters map[string]*cluster.Cluster
type Listeners map[string]*listener.Listener
type Endpoints map[string]*endpoint.ClusterLoadAssignment

type adapterServer struct {
	pb.UnimplementedAdapterServer
}

func listenerContents(listeners Listeners) []types.Resource {
	var r []types.Resource
	for _, l := range listeners {
		r = append(r, l)
	}
	return r
}

// MakeEndpoint creates a localhost endpoint on a given port.
func MakeEndpoint(clusterName string, address string, port uint32) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Policy: &endpoint.ClusterLoadAssignment_Policy{
			EndpointStaleAfter: &durationpb.Duration{Seconds: 5, Nanos: 0},
		},
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  address,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: port,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func MakeCluster(clusterName string, node string) *cluster.Cluster {
	edsSource := configSource(node)
	connectTimeout := 5 * time.Second
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(connectTimeout),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: edsSource,
		},
		DnsRefreshRate: &durationpb.Duration{
			Seconds: 5,
			Nanos:   0,
		},
	}
}

func MakeRoute(routeName, clusterName string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name:                routeName,
		InternalOnlyHeaders: []string{},
		VirtualHosts: []*route.VirtualHost{{
			Name:    routeName,
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				},
			}},
		}},
	}
}

// data source configuration
func configSource(clusterName string) *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = core.ApiVersion_V3
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       core.ApiVersion_V3,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: clusterName},
				},
			}},
		},
	}
	return source
}

func buildHttpConnectionManager() *hcm.HttpConnectionManager {
	// HTTP filter configuration.
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
		}},
	}
	return manager
}

func makeListener(listenerName string, address string, port uint32, filterChains []*listener.FilterChain) *listener.Listener {
	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains:   filterChains,
		TcpBacklogSize: &wrappers.UInt32Value{Value: 5},
	}
}

func MakeRouteHTTPListener(clusterName string, listenerName string, listenerAddress string, port uint32, route string) *listener.Listener {
	rdsSource := configSource(clusterName)
	routeSpecifier := &hcm.HttpConnectionManager_Rds{
		Rds: &hcm.Rds{
			ConfigSource:    rdsSource,
			RouteConfigName: route,
		},
	}

	manager := buildHttpConnectionManager()
	manager.RouteSpecifier = routeSpecifier

	pbst, err := ptypes.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	filterChains := []*listener.FilterChain{
		{
			Filters: []*listener.Filter{
				{
					Name: wellknown.HTTPConnectionManager,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				},
			},
		},
	}

	return makeListener(listenerName, listenerAddress, port, filterChains)
}

// MakeRuntime creates an RTDS layer with some fields.
func MakeRuntime(runtimeName string) *runtime.Runtime {
	return &runtime.Runtime{
		Name: runtimeName,
		Layer: &pstruct.Struct{
			Fields: map[string]*pstruct.Value{
				"field-0": {
					Kind: &pstruct.Value_NumberValue{NumberValue: 100},
				},
				"field-1": {
					Kind: &pstruct.Value_StringValue{StringValue: "foobar"},
				},
			},
		},
	}
}

func (a *adapterServer) SetState(ctx context.Context, request *pb.SetStateRequest) (response *pb.SetStateResponse, err error) {
	snapshot := cache.Snapshot{}
	if err != nil {
		return nil, err
	}

	for _, resourceReq := range request.Resources {
		resources := []types.Resource{}
		for _, name := range resourceReq.ResourceNames {
			switch resourceReq.TypeUrl {
			case TypeUrlCDS:
				newResource := MakeCluster(name, request.Node)
				resources = append(resources, newResource)
				snapshot.Resources[types.Cluster] = cache.NewResources(request.Version, resources)
			case TypeUrlLDS:
				address := fmt.Sprintf("https://%viscool.resources.com", name)
				newResource := makeListener(name, address, 11223, []*listener.FilterChain{})
				resources = append(resources, newResource)
				snapshot.Resources[types.Listener] = cache.NewResources(request.Version, resources)
			case TypeUrlRDS:
				newResource := MakeRoute(name, name)
				resources = append(resources, newResource)
				snapshot.Resources[types.Route] = cache.NewResources(request.Version, resources)
			case TypeUrlEDS:
				address := fmt.Sprintf("https://%viscool.endpoints.com", name)
				newResource := MakeEndpoint(name, address, 10000)
				resources = append(resources, newResource)
				snapshot.Resources[types.Endpoint] = cache.NewResources(request.Version, resources)
			}
		}
	}
	if err := xdsCache.SetSnapshot(context.Background(), request.Node, snapshot); err != nil {
		log.Printf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
	newSnapshot, err := xdsCache.GetSnapshot(request.Node)
	prettySnap, _ := json.Marshal(newSnapshot)
	fmt.Printf("new snapshot: \n%v\n\n", string(prettySnap))

	response = &pb.SetStateResponse{
		Success: true,
	}
	return response, nil

}

func (a *adapterServer) ClearState(ctx context.Context, req *pb.ClearStateRequest) (*pb.ClearStateResponse, error) {
	log.Printf("Clearing Cache")
	xdsCache.ClearSnapshot(req.Node)
	response := &pb.ClearStateResponse{
		Response: "All Clear",
	}
	return response, nil
}

func updateForType(res types.Resource) (uppedRes types.Resource) {
	switch v := res.(type) {
	case *cluster.Cluster:
		v.DnsRefreshRate.Seconds = v.DnsRefreshRate.Seconds + 5
	case *listener.Listener:
		v.TcpBacklogSize.Value = v.TcpBacklogSize.Value + 5
	case *route.RouteConfiguration:
		v.InternalOnlyHeaders = []string{"Testing"}
	case *endpoint.ClusterLoadAssignment:
		v.Policy.EndpointStaleAfter = &durationpb.Duration{Seconds: 10, Nanos: 0}
	default:
		fmt.Println("HUGH?", res.ProtoReflect().Type())
	}
	return res
}

func (a *adapterServer) UpdateResource(ctx context.Context, request *pb.ResourceRequest) (*pb.UpdateResourceResponse, error) {
	snapshot, err := xdsCache.GetSnapshot(request.Node)
	if err != nil {
		return nil, err
	}

	resources := []types.Resource{}

	for name, res := range snapshot.GetResources(request.TypeUrl) {
		if name != request.ResourceName {
			resources = append(resources, res)
		} else {
			uppedRes := updateForType(res)
			fmt.Println("uppedRes", res)
			resources = append(resources, uppedRes)
		}
	}

	switch request.TypeUrl {
	case TypeUrlCDS:
		snapshot.Resources[types.Cluster] = cache.NewResources(request.Version, resources)
	case TypeUrlLDS:
		snapshot.Resources[types.Listener] = cache.NewResources(request.Version, resources)
	case TypeUrlRDS:
		snapshot.Resources[types.Route] = cache.NewResources(request.Version, resources)
	case TypeUrlEDS:
		snapshot.Resources[types.Endpoint] = cache.NewResources(request.Version, resources)
	}

	if err := xdsCache.SetSnapshot(context.Background(), request.Node, snapshot); err != nil {
		return nil, err
	}

	response := &pb.UpdateResourceResponse{
		Success: true,
	}
	return response, nil
}

func (a *adapterServer) AddResource(ctx context.Context, request *pb.ResourceRequest) (*pb.AddResourceResponse, error) {
	snapshot, err := xdsCache.GetSnapshot(request.Node)
	if err != nil {
		return nil, err
	}

	resources := []types.Resource{}
	for _, res := range snapshot.GetResources(request.TypeUrl) {
		resources = append(resources, res)
	}

	switch request.TypeUrl {
	case TypeUrlCDS:
		newResource := MakeCluster(request.ResourceName, request.Node)
		resources = append(resources, newResource)
		snapshot.Resources[types.Cluster] = cache.NewResources(request.Version, resources)
	case TypeUrlLDS:
		address := fmt.Sprintf("https://%viscool.resources.com", request.ResourceName)
		newResource := makeListener(request.ResourceName, address, 11223, []*listener.FilterChain{})
		resources = append(resources, newResource)
		snapshot.Resources[types.Listener] = cache.NewResources(request.Version, resources)
	case TypeUrlRDS:
		newResource := MakeRoute(request.ResourceName, request.ResourceName)
		resources := append(resources, newResource)
		snapshot.Resources[types.Route] = cache.NewResources(request.Version, resources)
	case TypeUrlEDS:
		address := fmt.Sprintf("https://%viscool.endpoints.com", request.ResourceName)
		newResource := MakeEndpoint(request.ResourceName, address, 10000)
		resources := append(resources, newResource)
		snapshot.Resources[types.Endpoint] = cache.NewResources(request.Version, resources)
	}

	if err := xdsCache.SetSnapshot(context.Background(), request.Node, snapshot); err != nil {
		return nil, err
	}
	fmt.Println("Added Resource: ", request)
	response := &pb.AddResourceResponse{Success: true}
	return response, nil
}

func (a *adapterServer) RemoveResource(ctx context.Context, request *pb.ResourceRequest) (*pb.RemoveResourceResponse, error) {
	snapshot, err := xdsCache.GetSnapshot(request.Node)
	if err != nil {
		return nil, err
	}

	resources := []types.Resource{}
	for name, res := range snapshot.GetResources(request.TypeUrl) {
		if name != request.ResourceName {
			resources = append(resources, res)
		} else {
			fmt.Printf("Removing this resource: %v\n", name)
		}
	}

	switch request.TypeUrl {
	case TypeUrlCDS:
		snapshot.Resources[types.Cluster] = cache.NewResources(request.Version, resources)
	case TypeUrlLDS:
		snapshot.Resources[types.Listener] = cache.NewResources(request.Version, resources)
	case TypeUrlRDS:
		snapshot.Resources[types.Route] = cache.NewResources(request.Version, resources)
	case TypeUrlEDS:
		snapshot.Resources[types.Endpoint] = cache.NewResources(request.Version, resources)
	}

	if err := xdsCache.SetSnapshot(context.Background(), request.Node, snapshot); err != nil {
		return nil, err
	}
	response := &pb.RemoveResourceResponse{Success: true}
	return response, nil
}

func RunAdapter(port uint, cache cache.SnapshotCache) {
	xdsCache = cache
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterAdapterServer(s, &adapterServer{})
	log.Printf("Testsuite Adapter listening on %v\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Adapter failed to serve: %v", err)
	}
}
