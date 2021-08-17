package runner

import (
	"context"
	"fmt"
	"io"
	"time"

	cluster_service "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
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

type XDSMessages struct {
	Responses chan string
	Errors    chan error
	Done      chan bool
}

type Runner struct {
	Adapter *ClientConfig
	Target  *ClientConfig
	CDS     *XDSMessages
}

func NewRunner() *Runner {
	return &Runner{
		Adapter: &ClientConfig{},
		Target:  &ClientConfig{},
	}
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
func (r *Runner) CDSAckAck(dreq *discovery.DiscoveryRequest, dres chan<- *discovery.DiscoveryResponse, errors chan<- error, done chan<- bool) {
	c := cluster_service.NewClusterDiscoveryServiceClient(r.Target.Conn)
	stream, err := c.StreamClusters(context.Background())
	if err != nil {
		errors <- err
		return
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(dres)
				close(errors)
				return
			}
			if err != nil {
				errors <- err
				close(dres)
				close(errors)
				return
			}
			dres <- in
		}
	}()
	if err := stream.Send(dreq); err != nil {
		errors <- err
	}
	stream.CloseSend()
	<-waitc
}
