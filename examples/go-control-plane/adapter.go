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
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"google.golang.org/grpc"
)

var (
	xdsCache  cache.SnapshotCache
	localhost = "127.0.0.1"
	XDSCache cache.LinearCache
)

const (
	TypeUrlLDS = "type.googleapis.com/envoy.config.listener.v3.Listener"
	TypeUrlCDS = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	TypeUrlRDS = "type.googleapis.com/envoy.config.route.v3.RouteConfiguration"
	TypeUrlEDS = "type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment"
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
	}
}

func MakeRoute(routeName, clusterName string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeName,
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
		FilterChains: filterChains,
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

func (a *adapterServer) SetState(ctx context.Context, state *pb.Snapshot) (response *pb.SetStateResponse, err error) {

	clusters := make([]types.Resource, len(state.Clusters.Items))
	for i, cluster := range state.Clusters.Items {
		clusters[i] = MakeCluster(cluster.Name, state.Node)
	}

	endpoints := make([]types.Resource, len(state.Endpoints.Items))
	for i, endpoint := range state.Endpoints.Items {
		endpoints[i] = MakeEndpoint(endpoint.Cluster, endpoint.Address, uint32(10000+i))
	}

	routes := make([]types.Resource, len(state.Routes.Items))
	for i, route := range state.Routes.Items {
		cluster := state.Clusters.Items[i]
		routes[i] = MakeRoute(route.Name, cluster.Name)
	}

	listeners := make([]types.Resource, len(state.Listeners.Items))
	for i, listener := range state.Listeners.Items {
		port := uint32(11000 + i)
		route := state.Routes.Items[i]
		listeners[i] = MakeRouteHTTPListener(state.Node, listener.Name, listener.Address, port, route.Name)
	}

	runtimes := make([]types.Resource, len(state.Runtimes.Items))
	for i, runtime := range state.Runtimes.Items {
		runtimes[i] = MakeRuntime(runtime.Name)
	}

	snapshot, err := cache.NewSnapshot(
		state.Version,
		map[resource.Type][]types.Resource{
			resource.EndpointType: endpoints,
			resource.ClusterType: clusters,
			resource.RouteType: routes,
			resource.ListenerType: listeners,
			resource.RuntimeType: runtimes,
			resource.SecretType: {},
		},
	)
	if err != nil {
		log.Printf("Error creating snapshot: %v", err)
	}
	if err = snapshot.Consistent(); err != nil {
		log.Printf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		os.Exit(1)
	}

	// // Add the snapshot to the cache
	if err := xdsCache.SetSnapshot(context.Background(), state.Node, snapshot); err != nil {
		log.Printf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}

	newSnapshot, err := xdsCache.GetSnapshot(state.Node)
	prettySnap, _ := json.Marshal(newSnapshot)
	fmt.Printf("new snapshot: \n%v\n\n", string(prettySnap))
	response = &pb.SetStateResponse{
		Message: "Success",
	}
	return response, nil
}

func (a *adapterServer) UpdateState(ctx context.Context, state *pb.Snapshot) (*pb.UpdateStateResponse, error) {
	response, err := a.SetState(ctx, state)
	if err != nil {
		fmt.Printf("Error setting state: %v", err)
		return nil, err
	}
	updateResponse := &pb.UpdateStateResponse{
		Message: response.Message,
	}
	return updateResponse, err
}

func (a *adapterServer) SetResources(ctx context.Context, req *pb.SetResourcesRequest) (response *pb.SetResourcesResponse, err error) {
	XDSCache = *cache.NewLinearCache(req.TypeURL)
	resources := makeXdsResources(req.TypeURL, req.Resources)

	XDSCache.SetResources(resources)

	// TODO find better response, confirm that it worked basically.
	response = &pb.SetResourcesResponse{
		Version: "1",
		Message: "it worked, but this a dummy message.",
	}
	return response, nil
}

func (a *adapterServer) UpdateResource (ctx context.Context, request *pb.ResourceRequest) (*pb.UpdateResourceResponse, error) {
	linear := cache.NewLinearCache(request.TypeURL)
	fmt.Printf("The resources: %v\n", linear.GetResources())

	return nil, fmt.Errorf("Zach is very cool!")
}

func (a *adapterServer) ClearState(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	log.Printf("Clearing Cache")
	xdsCache.ClearSnapshot(req.Node)
	response := &pb.ClearResponse{
		Response: "All Clear",
	}
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

func makeXdsResources(typeURL string, resources []*pb.Resource) map[string]types.Resource {
	xdsResources := make(map[string]types.Resource)

	switch typeURL {
	case TypeUrlCDS:
		for _, resource := range resources {
			xdsResources[resource.Name] = MakeCluster(resource.Name,  "test-id")
		}
	case TypeUrlLDS:
		for _, resource := range resources {
			xdsResources[resource.Name] = MakeRouteHTTPListener("test-id", resource.Name, "gagagagagaga.com", 180000, resource.Name)
		}
	case TypeUrlRDS:
		for _, resource := range resources {
			xdsResources[resource.Name] = MakeRoute(resource.Name, resource.Name)
		}
	case TypeUrlEDS:
		for _, resource := range resources {
			xdsResources[resource.Name] = MakeEndpoint(resource.Name, resource.Name, 18000)
		}
	}
	return xdsResources
}
	// endpoints := make([]types.Resource, len(state.Endpoints.Items))
	// for i, endpoint := range state.Endpoints.Items {
	// 	endpoints[i] = MakeEndpoint(endpoint.Cluster, endpoint.Address, uint32(10000+i))
	// }

	// routes := make([]types.Resource, len(state.Routes.Items))
	// for i, route := range state.Routes.Items {
	// 	cluster := state.Clusters.Items[i]
	// 	routes[i] = MakeRoute(route.Name, cluster.Name)
	// }

	// listeners := make([]types.Resource, len(state.Listeners.Items))
	// for i, listener := range state.Listeners.Items {
	// 	port := uint32(11000 + i)
	// 	route := state.Routes.Items[i]
	// 	listeners[i] = MakeRouteHTTPListener(state.Node, listener.Name, listener.Address, port, route.Name)
	// }

	// runtimes := make([]types.Resource, len(state.Runtimes.Items))
	// for i, runtime := range state.Runtimes.Items {
	// 	runtimes[i] = MakeRuntime(runtime.Name)
	// }
