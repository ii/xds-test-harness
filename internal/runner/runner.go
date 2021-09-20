package runner

import (
	// "context"
	"context"
	"errors"
	"fmt"
	"sync"

	"io"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"github.com/ii/xds-test-harness/internal/parser"
	pb "github.com/zachmandeville/tester-prototype/api/adapter"
	"google.golang.org/grpc"
	// "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	StartState *pb.Snapshot
	FinalResponse *discovery.DiscoveryResponse
}

type CDS struct {
	Req   chan *discovery.DiscoveryRequest
	Res   chan *discovery.DiscoveryResponse
	Err   chan error
	Done  chan bool
	Cache struct {
		Requests  []*discovery.DiscoveryRequest
		Responses []*discovery.DiscoveryResponse
	}
}

type Runner struct {
	Adapter *ClientConfig
	Target  *ClientConfig
	NodeID  string
	Cache   *Cache
	CDS     *CDS
}

func NewRunner() *Runner {
	return &Runner{
		Adapter: &ClientConfig{},
		Target:  &ClientConfig{},
		NodeID:  "",
		Cache:   &Cache{},
		CDS:     &CDS{},
	}
}

func (r *Runner) NewWildcardCDSRequest() *discovery.DiscoveryRequest {
	return &discovery.DiscoveryRequest{
		VersionInfo: "",
		Node: &core.Node{
			Id: r.NodeID,
		},
		ResourceNames: []string{},
		TypeUrl:       "type.googleapis.com/envoy.config.cluster.v3.Cluster",
	}
}

func (r *Runner) AckCDS(initReq *discovery.DiscoveryRequest) {

	log.Debug().Msg("Sending First Discovery Request")
	r.CDS.Req <- initReq
	r.CDS.Cache.Requests = append(r.CDS.Cache.Requests, initReq)

	for {
		select {
		case res := <-r.CDS.Res:
			r.CDS.Cache.Responses = append(r.CDS.Cache.Responses, res)
			ack, err := NewCDSAckFromResponse(res)
			if err != nil {
				log.Err(err).Msg("Error creating Ack Request")
			}
			log.Debug().Msgf("Got response, created ack: %v\n", ack)
			r.CDS.Req <- ack
	        r.CDS.Cache.Requests = append(r.CDS.Cache.Requests, ack)
		case <-r.CDS.Done:
			log.Debug().Msg("Received Done signal, shutting down request channel")
			close(r.CDS.Req)
			return
		}
	}
}

func NewCDSAckFromResponse(res *discovery.DiscoveryResponse) (*discovery.DiscoveryRequest, error) {
	response, err := parser.ParseDiscoveryResponse(res)
	if err != nil {
		err := fmt.Errorf("error parsing dres for acking: %v", err)
		return nil, err
	}
	clusters := []string{}
	for _, cluster := range response.Resources {
		clusters = append(clusters, cluster.Name)
	}
	request := &discovery.DiscoveryRequest{
		VersionInfo:   response.VersionInfo,
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
	log.Debug().
		Msgf("Runner connected to %v", server)
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
	conn, err := connectViaGRPC(r.Adapter, "adapter")
	if err != nil {
		return err
	}
	r.Adapter.Conn = conn
	return nil
}

func (r *Runner) CDSStream() error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer close(r.CDS.Err)
	client := cds.NewClusterDiscoveryServiceClient(r.Target.Conn)
	stream, err := client.StreamClusters(ctx)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot start CDS stream %v. error: %v", stream, err))
		log.Debug().
			Err(err).
			Msg("")
		r.CDS.Err <- err
	}
	var wg sync.WaitGroup
	go func() {
		for {
			wg.Add(1)
			in, err := stream.Recv()
			if err == io.EOF {
				log.Debug().
					Msg("No more discovery responses for CDS stream")
				close(r.CDS.Res)
				return
			}
			if err != nil {
				log.Err(err).Msg("error receiving responses on CDS stream")
				r.CDS.Err <- err
			}
			log.Debug().
				Msgf("Received discovery response: %v", in)
			r.CDS.Res <- in
		}
	}()

	for req := range r.CDS.Req {
		log.Debug().
			Msgf("Received req from channel: %v", req)
		if err := stream.Send(req); err != nil {
			log.Err(err)
			r.CDS.Err <- err
		}
	}
	stream.CloseSend()
	wg.Wait()
	return nil
}
