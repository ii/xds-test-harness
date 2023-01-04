package ii.nz.xdstestharness.adapter;

import com.google.common.collect.ImmutableList;
import com.google.protobuf.Any;
import com.google.protobuf.InvalidProtocolBufferException;
import io.envoyproxy.controlplane.cache.Resources;
import io.envoyproxy.controlplane.cache.TestResources;
import io.envoyproxy.controlplane.cache.v3.SimpleCache;
import io.envoyproxy.controlplane.cache.v3.Snapshot;
import io.envoyproxy.envoy.config.cluster.v3.Cluster;
import io.envoyproxy.envoy.config.core.v3.ApiVersion;
import io.envoyproxy.envoy.config.endpoint.v3.ClusterLoadAssignment;
import io.envoyproxy.envoy.config.listener.v3.Listener;
import io.envoyproxy.envoy.config.route.v3.RouteConfiguration;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import nz.ii.xdstestharness.adapter.AdapterGrpc;
import nz.ii.xdstestharness.adapter.AdapterProto;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;

public class AdapterServer {
    static final String TypeURLCDS = "type.googleapis.com/envoy.config.cluster.v3.Cluster";
    static final String TypeURLLDS = "type.googleapis.com/envoy.config.listener.v3.Listener";
    static final String TypeURLRDS = "type.googleapis.com/envoy.config.route.v3.RouteConfiguration";
    static final String TypeURLEDS = "type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment";
    private final int port;
    private static SimpleCache<String> cache;

    private final Server server;

    public AdapterServer(int port, SimpleCache<String> cache) throws IOException {
        this(ServerBuilder.forPort(port), port, cache);
    }

    public AdapterServer(ServerBuilder<?> serverBuilder, int port, SimpleCache<String> cache) {
        this.port = port;
        this.cache = cache;
        server = serverBuilder.addService(new AdapterService())
                .build();
    }

    public void start() throws IOException {
        server.start();
        Runtime.getRuntime().addShutdownHook(new Thread() {
            @Override
            public void run() {
                System.err.println("*** shutting down gRPC adapter server since JVM is shutting down");
                try {
                    AdapterServer.this.stop();
                } catch(InterruptedException e) {
                    e.printStackTrace(System.err);
                }
                System.err.println("*** server shut down");

        }
        });
    }

    public void stop() throws InterruptedException {
        if (server != null) {
            server.shutdown().awaitTermination(30, TimeUnit.SECONDS);
        }
    }

    public void blockUntilShutdown() throws InterruptedException {
        if (server != null) {
            server.awaitTermination();
        }
    }

    public static void main(String[] args) throws Exception {
        AdapterServer server = new AdapterServer(8980, cache);
        server.start();
        System.out.println("Server running on port: "+8980);
        server.blockUntilShutdown();
    }

    private static class AdapterService extends AdapterGrpc.AdapterImplBase {
        AdapterService() { }

        private Cluster resourceToCluster(Any resource) {
            Cluster c;
            try {
                c = resource.unpack(Cluster.class);
            } catch (InvalidProtocolBufferException e) {
                throw new RuntimeException(e);
            }
            return TestResources.createCluster( c.getName());
        }

        private Listener resourceToListener(Any resource) {
            Listener l;
            ApiVersion rdsVersion = ApiVersion.V3;
            try {
                l = resource.unpack(Listener.class);
            } catch (InvalidProtocolBufferException e) {
                throw new RuntimeException(e);
            }
            return TestResources.createListener(
                    false,
                    false,
                    rdsVersion,
                    rdsVersion,
                    l.getName(),
                    1234,
                    l.getName());
        }

        private ClusterLoadAssignment resourceToEndpoint(Any resource) {
            ClusterLoadAssignment cla;
            try {
                cla = resource.unpack(ClusterLoadAssignment.class);
            } catch (InvalidProtocolBufferException e) {
                throw new RuntimeException(e);
            }
            return TestResources.createEndpoint(cla.getClusterName(),1234);
        }

       private RouteConfiguration resourceToRoute(Any resource) {
            RouteConfiguration r;
            try {
                r = resource.unpack(RouteConfiguration.class);
            } catch (InvalidProtocolBufferException e) {
                throw new RuntimeException(e);
            }
            return TestResources.createRoute(r.getName(),r.getName());
        }

        public void setState(AdapterProto.SetStateRequest request, StreamObserver<AdapterProto.SetStateResponse> responseObserver) {
            String node = request.getNode();
            String version = request.getVersion();
            List<com.google.protobuf.Any> resources = request.getResourcesList();
            List<Cluster> clusters = new ArrayList<Cluster>();
            List<Listener> listeners = new ArrayList<Listener>();
            List<ClusterLoadAssignment> endpoints = new ArrayList<ClusterLoadAssignment>();
            List<RouteConfiguration> routes = new ArrayList<RouteConfiguration>();
            for (Any r : resources) {
                String typeURL = r.getTypeUrl();
                switch(typeURL) {
                    case TypeURLCDS:
                        clusters.add(resourceToCluster(r));
                        break;
                    case TypeURLLDS:
                        listeners.add(resourceToListener(r));
                        break;
                    case TypeURLEDS:
                        endpoints.add(resourceToEndpoint(r));
                        break;
                    case TypeURLRDS:
                        routes.add(resourceToRoute(r));
                        break;
                    default:
                        System.out.println("New type url we weren't expecting: "+typeURL);
                        break;
                }
            }

            cache.setSnapshot(
                    node,
                    Snapshot.create(
                            clusters,
                            endpoints,
                            listeners,
                            routes,
                            ImmutableList.of(),
                            version));

            AdapterProto.SetStateResponse response =
                    AdapterProto.SetStateResponse.newBuilder().setSuccess(true).build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }

        public void clearState(AdapterProto.clearStateRequest request, StreamObserver<AdapterProto.clearStateResponse> responseObserver) {
            System.out.println("Clearing the cache");
            cache.clearSnapshot(request.getNode());
            AdapterProto.clearStateResponse response =
                    AdapterProto.clearStateResponse.
                            newBuilder().
                            setResponse("Success").
                            build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }

        public void updateResource(AdapterProto.ResourceRequest request, StreamObserver<AdapterProto.UpdateResourceResponse> responseObserver) {
            System.out.println("getting update request!");
            io.envoyproxy.controlplane.cache.Snapshot state = cache.getSnapshot(request.getNode());

            List<Cluster> clusters = new ArrayList<Cluster>();
            Map<String, Cluster> stateClusters = (Map<String, Cluster>) state.resources(Resources.ResourceType.CLUSTER);
            for (Map.Entry<String,Cluster> entry : stateClusters.entrySet()) {
                if ((request.getTypeUrl().equals(TypeURLCDS))
                        && request.getResourceName().equals(entry.getKey())) {
                    clusters.add(TestResources.createCluster(entry.getKey(), "127.0.0.1", 8888, Cluster.DiscoveryType.STATIC));
                } else {
                    clusters.add(entry.getValue());
                }
            };

            List<ClusterLoadAssignment> endpoints = new ArrayList<ClusterLoadAssignment>();
            Map<String, ClusterLoadAssignment> stateEndpoints = (Map<String, ClusterLoadAssignment>) state.resources((Resources.ResourceType.ENDPOINT));
            for (Map.Entry<String,ClusterLoadAssignment> entry : stateEndpoints.entrySet()) {
                if (request.getTypeUrl().equals(TypeURLEDS)
                    && entry.getKey().equals(request.getResourceName())) {
                    endpoints.add(TestResources.createEndpoint(entry.getKey(),8888));
                } else {
                    endpoints.add(entry.getValue());
                }
            }

            List<Listener> listeners = new ArrayList<Listener>();
            Map<String, Listener> stateListeners = (Map<String, Listener>) state.resources((Resources.ResourceType.LISTENER));
            for (Map.Entry<String,Listener> entry : stateListeners.entrySet()) {
                if (request.getTypeUrl().equals(TypeURLLDS)
                        && entry.getKey().equals(request.getResourceName())) {
                    listeners.add( TestResources.createListener(
                            false,
                            false,
                            ApiVersion.V3,
                            ApiVersion.V3,
                            entry.getKey(),
                            8888,
                            entry.getKey()));
                } else {
                    listeners.add(entry.getValue());
                }
            }

            List<RouteConfiguration> routes = new ArrayList<RouteConfiguration>();
            Map<String,RouteConfiguration> stateRoutes = (Map<String,RouteConfiguration>) state.resources(((Resources.ResourceType.ROUTE)));
            for (Map.Entry<String,RouteConfiguration> entry : stateRoutes.entrySet()) {
                if (request.getTypeUrl().equals(TypeURLRDS)
                        && request.getResourceName().equals(entry.getKey())) {
                    routes.add(TestResources.createRoute(entry.getKey(), entry.getKey()+"1"));
                } else {
                    routes.add(entry.getValue());
                }
            }

            cache.setSnapshot(
                    request.getNode(),
                    Snapshot.create(
                            clusters,
                            endpoints,
                            listeners,
                            routes,
                            ImmutableList.of(),
                            request.getVersion()
                    ));

            AdapterProto.UpdateResourceResponse response =
                    AdapterProto.UpdateResourceResponse.newBuilder().setSuccess(true).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }

        public void addResource(AdapterProto.ResourceRequest request, StreamObserver<AdapterProto.AddResourceResponse> responseObserver) {
            System.out.println("getting addResource request!");
            io.envoyproxy.controlplane.cache.Snapshot state = cache.getSnapshot(request.getNode());

            List<Cluster> clusters = new ArrayList<Cluster>();
            Map<String, Cluster> stateClusters = (Map<String, Cluster>) state.resources(Resources.ResourceType.CLUSTER);
            for (Map.Entry<String,Cluster> entry : stateClusters.entrySet()) {
                clusters.add(entry.getValue());
            }

            List<ClusterLoadAssignment> endpoints = new ArrayList<ClusterLoadAssignment>();
            Map<String, ClusterLoadAssignment> stateEndpoints = (Map<String, ClusterLoadAssignment>) state.resources((Resources.ResourceType.ENDPOINT));
            for (Map.Entry<String,ClusterLoadAssignment> entry : stateEndpoints.entrySet()) {
                endpoints.add(entry.getValue());
            }

            List<Listener> listeners = new ArrayList<Listener>();
            Map<String, Listener> stateListeners = (Map<String, Listener>) state.resources((Resources.ResourceType.LISTENER));
            for (Map.Entry<String,Listener> entry : stateListeners.entrySet()) {
                listeners.add(entry.getValue());
            }

            List<RouteConfiguration> routes = new ArrayList<RouteConfiguration>();
            Map<String,RouteConfiguration> stateRoutes = (Map<String,RouteConfiguration>) state.resources(((Resources.ResourceType.ROUTE)));
            for (Map.Entry<String,RouteConfiguration> entry : stateRoutes.entrySet()) {
                routes.add(entry.getValue());
            }
            switch(request.getTypeUrl()) {
                case TypeURLCDS:
                    clusters.add(TestResources.createCluster(request.getResourceName()));
                    break;
                case TypeURLLDS:
                    listeners.add(TestResources.createListener(
                            false,
                            false,
                            ApiVersion.V3,
                            ApiVersion.V3,
                            request.getResourceName(),
                            1234,
                            request.getResourceName()));
                    break;
                case TypeURLEDS:
                    endpoints.add(TestResources.createEndpoint(request.getResourceName(), 1234));
                    break;
                case TypeURLRDS:
                    routes.add(TestResources.createRoute(request.getResourceName(), request.getResourceName()));
                    break;
                default:
                    System.out.println("New type url we weren't expecting: "+request.getTypeUrl());
                    break;
            }

            cache.setSnapshot(
                    request.getNode(),
                    Snapshot.create(
                            clusters,
                            endpoints,
                            listeners,
                            routes,
                            ImmutableList.of(),
                            request.getVersion()
                    ));
            AdapterProto.AddResourceResponse response =
                    AdapterProto.AddResourceResponse.newBuilder().setSuccess(true).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }

        public void removeResource(AdapterProto.ResourceRequest request, StreamObserver<AdapterProto.RemoveResourceResponse> responseObserver) {
            System.out.println("getting removeResource request!");
            io.envoyproxy.controlplane.cache.Snapshot state = cache.getSnapshot(request.getNode());

            List<Cluster> clusters = new ArrayList<Cluster>();
            Map<String, Cluster> stateClusters = (Map<String, Cluster>) state.resources(Resources.ResourceType.CLUSTER);
            for (Map.Entry<String,Cluster> entry : stateClusters.entrySet()) {
                if ((request.getTypeUrl().equals(TypeURLCDS))
                        && request.getResourceName().equals(entry.getKey())) {
                    System.out.println("Removing Resource "+entry.getKey()+" From "+request.getTypeUrl());
                } else {
                    clusters.add(entry.getValue());
                }
            };

            List<ClusterLoadAssignment> endpoints = new ArrayList<ClusterLoadAssignment>();
            Map<String, ClusterLoadAssignment> stateEndpoints = (Map<String, ClusterLoadAssignment>) state.resources((Resources.ResourceType.ENDPOINT));
            for (Map.Entry<String,ClusterLoadAssignment> entry : stateEndpoints.entrySet()) {
                if (request.getTypeUrl().equals(TypeURLEDS)
                        && entry.getKey().equals(request.getResourceName())) {
                    System.out.println("Removing Resource "+entry.getKey()+" From "+request.getTypeUrl());
                } else {
                    endpoints.add(entry.getValue());
                }
            }

            List<Listener> listeners = new ArrayList<Listener>();
            Map<String, Listener> stateListeners = (Map<String, Listener>) state.resources((Resources.ResourceType.LISTENER));
            for (Map.Entry<String,Listener> entry : stateListeners.entrySet()) {
                if (request.getTypeUrl().equals(TypeURLLDS)
                        && entry.getKey().equals(request.getResourceName())) {
                    System.out.println("Removing Resource "+entry.getKey()+" From "+request.getTypeUrl());
                } else {
                    listeners.add(entry.getValue());
                }
            }

            List<RouteConfiguration> routes = new ArrayList<RouteConfiguration>();
            Map<String,RouteConfiguration> stateRoutes = (Map<String,RouteConfiguration>) state.resources(((Resources.ResourceType.ROUTE)));
            for (Map.Entry<String,RouteConfiguration> entry : stateRoutes.entrySet()) {
                if (request.getTypeUrl().equals(TypeURLRDS)
                        && request.getResourceName().equals(entry.getKey())) {
                    System.out.println("Removing Resource "+entry.getKey()+" From "+request.getTypeUrl());
                } else {
                    routes.add(entry.getValue());
                }
            }

            cache.setSnapshot(
                    request.getNode(),
                    Snapshot.create(
                            clusters,
                            endpoints,
                            listeners,
                            routes,
                            ImmutableList.of(),
                            request.getVersion()
                    ));

            AdapterProto.RemoveResourceResponse response =
            AdapterProto.RemoveResourceResponse.newBuilder().setSuccess(true).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
    }
}
