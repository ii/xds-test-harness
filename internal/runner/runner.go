package runner

import (
	"context"
	"fmt"
	"io"
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/ii/xds-test-harness/internal/parser"
	"google.golang.org/grpc"
	pb "github.com/zachmandeville/tester-prototype/api/adapter"
)

var (
	opts []grpc.DialOption = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}
)

type ClientConfig struct {
	Port string
	Conn *grpc.ClientConn
}

type Cache struct {
	Response *discovery.DiscoveryResponse
	Snapshot *pb.Snapshot
}

type CDSCache struct {
	Responses []*discovery.DiscoveryResponse
	Stream cluster_service.ClusterDiscoveryService_StreamClustersClient
}


type Runner struct {
	Adapter *ClientConfig
	Target  *ClientConfig
	NodeID  string
	Cache   *Cache
	CDSCache *CDSCache
}

func NewRunner() *Runner {
	return &Runner{
		Adapter: &ClientConfig{},
		Target:  &ClientConfig{},
		Cache: &Cache{},
		NodeID: "",
		CDSCache: &CDSCache{},
	}
}

func NewWildcardCDSRequest (nodeID string) *discovery.DiscoveryRequest {
	return &discovery.DiscoveryRequest{
		VersionInfo: "",
		Node: &envoy_config_core_v3.Node{
			Id: nodeID,
		},
		ResourceNames: []string{},
		TypeUrl:       "type.googleapis.com/envoy.config.cluster.v3.Cluster",
	}
}

func NewCDSAckRequestFromResponse(node string, res *discovery.DiscoveryResponse) (*discovery.DiscoveryRequest, error) {
	response, err:= parser.ParseDiscoveryResponse(res)
	if err != nil {
		err := fmt.Errorf("error parsing dres for acking: %v", err)
		return nil, err
	}
	clusters := []string{}
	for _, cluster := range response.Resources {
		clusters = append(clusters, cluster.Name)
	}
	request := &discovery.DiscoveryRequest{
		VersionInfo: response.VersionInfo,
		Node: &envoy_config_core_v3.Node{
			Id: node,
		},
		ResourceNames: clusters,
		TypeUrl:       "type.googleapis.com/envoy.config.cluster.v3.Cluster",
		ResponseNonce: response.Nonce,
	}
	return request, nil
}

func connectViaGRPC(client *ClientConfig, server string) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.Dial(client.Port, opts...)
	if err != nil {
		err = fmt.Errorf("Cannot connect at %v: %v\n", client.Port, err)
		return nil, err
	}
	return conn, nil
}

func (r *Runner) ConnectToTarget(address string) error {
	r.Target.Port = address
	conn, err := connectViaGRPC(r.Target, "target")
	if err != nil {
		return err
	}
	r.Target.Conn = conn
	return nil
}

func (r *Runner) ConnectToAdapter(address string) error {
	r.Adapter.Port = address
	fmt.Println("adapter: ", r.Adapter.Port)
	conn, err := connectViaGRPC(r.Adapter, "adapter")
	if err != nil {
		return err
	}
	r.Adapter.Conn = conn
	return nil
}

// starts stream with CDS with given discovery request, dreq.
// sends discovery response to r.dRes channel,
// sends any errors to r.channels.errors
// closes strema after acking successful dResponse and sends message on Done channel
func (r *Runner) CDSAckAck(dreq <-chan *discovery.DiscoveryRequest, dres chan<- *discovery.DiscoveryResponse, errors chan<- error, done chan<- bool) {
	c := cluster_service.NewClusterDiscoveryServiceClient(r.Target.Conn)
	stream, err := c.StreamClusters(context.Background())
	r.CDSCache.Stream = stream
	if err != nil {
		errors <- err
		return
	}
	go func() {
		for {
			in, err := r.CDSCache.Stream.Recv()
			if err == io.EOF {
				done <- true
				close(dres)
				close(errors)
				close(done)
				return
			}
			if err != nil {
				err = fmt.Errorf("Error receiving from stream: %v\n", err)
				errors <- err
				close(dres)
				close(errors)
				return
			}
			dres <- in
		}
	}()
	for req := range dreq {
		if err := stream.Send(req); err != nil {
			err = fmt.Errorf("Error sending discovery request: %v", err)
		}
	}
	done <- true
}
