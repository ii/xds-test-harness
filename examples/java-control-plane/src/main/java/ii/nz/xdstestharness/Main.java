package ii.nz.xdstestharness;

import java.io.IOException;

import ii.nz.xdstestharness.adapter.AdapterServer;
import io.envoyproxy.controlplane.cache.v3.SimpleCache;
import io.envoyproxy.controlplane.server.V3DiscoveryServer;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.netty.NettyServerBuilder;

public class Main {
    private static final String GROUP = "test-id";
    private static final int xdsPort = 19000;

    public static void main(String[] args) throws IOException, InterruptedException {
        SimpleCache<String> cache = new SimpleCache<>(node->GROUP);
        V3DiscoveryServer v3DiscoveryServer = new V3DiscoveryServer(cache);

        ServerBuilder builder =
                NettyServerBuilder.forPort(xdsPort)
                        .addService(v3DiscoveryServer.getAggregatedDiscoveryServiceImpl())
                        .addService(v3DiscoveryServer.getClusterDiscoveryServiceImpl())
                        .addService(v3DiscoveryServer.getEndpointDiscoveryServiceImpl())
                        .addService(v3DiscoveryServer.getListenerDiscoveryServiceImpl())
                        .addService(v3DiscoveryServer.getRouteDiscoveryServiceImpl());

        Server server = builder.build();
        AdapterServer adapter = new AdapterServer(8980, cache);

        new Thread(()->{
            try {
                server.start();
                System.out.println("Server has started on port " + server.getPort());
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
            try {
                server.awaitTermination();
            } catch (InterruptedException e) {
                throw new RuntimeException(e);
            }
        }).start();

        new Thread(()->{
            try {
                adapter.start();
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
            System.out.println("Adapter Server running on port: "+8980);
            try {
                adapter.blockUntilShutdown();
            } catch (InterruptedException e) {
                throw new RuntimeException(e);
            }
        }).start();
    }
}
