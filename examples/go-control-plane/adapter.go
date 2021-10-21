package example

import (
	"context"
	"fmt"
	"log"
	"encoding/json"
	"net"
	"os"
	"time"
	"google.golang.org/grpc"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	runtime "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

var (
	xdsCache cache.SnapshotCache
	localhost = "127.0.0.1"
)

type Clusters  map[string]*cluster.Cluster
type Listeners map[string]*listener.Listener
type Endpoints map[string]*endpoint.ClusterLoadAssignment

type adapterServer struct {
	pb.UnimplementedAdapterServer
}

func listenerContents(listeners Listeners) []types.Resource {
	var r []types.Resource
	for _, l := range listeners {
		r = append(r,l)
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

func (a *adapterServer) SetState (ctx context.Context, state *pb.Snapshot) (response *pb.SetStateResponse, err error) {

	numClusters := len(state.Clusters.Items)
	clusters := make([]types.Resource, numClusters)
	for i := 0; i < numClusters; i++ {
		cluster := state.Clusters.Items[i]
		clusters[i] = MakeCluster(cluster.Name, state.Node)

	}

	numEndpoints := len(state.Endpoints.Items)
	endpoints := make([]types.Resource, numEndpoints)
	for i := 0; i < numEndpoints; i++ {
  	    endpoint:= state.Endpoints.Items[i]
		endpoints[i] = MakeEndpoint(endpoint.Cluster, endpoint.Address, uint32(10000+i))
	}

	numRoutes := len(state.Routes.Items)
	routes := make([]types.Resource, numRoutes)
	// TODO grab the cluster from the route itself, by updating api?
	for i := 0; i < numRoutes; i++ {
		route := state.Routes.Items[i]
		cluster := state.Clusters.Items[i]
		routes[i] = MakeRoute(route.Name, cluster.Name)
	}

	numListeners := len(state.Listeners.Items)
	listeners := make([]types.Resource, numListeners)
	for i := 0; i < numListeners; i++ {
		listener := state.Listeners.Items[i]
		port := uint32(11000 + i)
		route := state.Routes.Items[i]
		listeners[i] = MakeRouteHTTPListener(state.Node, listener.Name, listener.Address, port, route.Name)
	}

	numRuntimes := len(state.Runtimes.Items)
	runtimes := make([]types.Resource, numRuntimes)
	for i := 0; i < numRuntimes; i++ {
		runtime := state.Runtimes.Items[i]
		runtimes[i] = MakeRuntime(runtime.Name)
	}

	snapshot := cache.NewSnapshot(
		state.Version,
		endpoints,
		clusters,
		routes,
		listeners,
		runtimes,
		[]types.Resource{}, // secrets
	)
	if err = snapshot.Consistent(); err != nil {
		log.Printf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		os.Exit(1)
	}

	// // Add the snapshot to the cache
	if err := xdsCache.SetSnapshot(state.Node, snapshot); err != nil {
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
