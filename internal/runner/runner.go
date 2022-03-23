package runner

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	pb "github.com/ii/xds-test-harness/api/adapter"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	opts []grpc.DialOption = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 10),
	}
)

type ClientConfig struct {
	Port string
	Conn *grpc.ClientConn
}

type Cache struct {
	StartState     *pb.Snapshot
	StateSnapshots []*pb.Snapshot
	FinalResponse  *discovery.DiscoveryResponse
}

type Runner struct {
	Adapter    *ClientConfig
	Target     *ClientConfig
	NodeID     string
	Cache      *Cache
	Aggregated bool
	Service    *XDSService
	SubscribeRequest *discovery.DiscoveryRequest
}

func FreshRunner(current ...*Runner) *Runner {
	var (
		adapter    = &ClientConfig{}
		target     = &ClientConfig{}
		nodeID     = ""
		aggregated = false
	)

	if len(current) > 0 {
		adapter = current[0].Adapter
		target = current[0].Target
		nodeID = current[0].NodeID
		aggregated = current[0].Aggregated

	}

	return &Runner{
		Adapter:    adapter,
		Target:     target,
		NodeID:     nodeID,
		Cache:      &Cache{},
		Service:    &XDSService{},
		Aggregated: aggregated,
	}
}

func (r *Runner) ConnectClient(server, address string) error {
	var client *ClientConfig
	if server == "target" {
		client = r.Target
	}
	if server == "adapter" {
		client = r.Adapter
	}
	if strings.HasPrefix(address, ":") {
		client.Port = address
	} else {
		client.Port = ":" + address
	}
	conn, err := connectViaGRPC(client, server)
	if err != nil {
		return err
	}
	client.Conn = conn
	return nil
}

func (r *Runner) Ack(service *XDSService) {
	service.Channels.Req <- r.SubscribeRequest
	service.Cache.Requests = append(service.Cache.Requests, r.SubscribeRequest)
	for {
		select {
		case res := <-service.Channels.Res:
			service.Cache.Responses = append(service.Cache.Responses, res)
			ack := newAckFromResponse(res, r.SubscribeRequest)
			log.Debug().
				Msgf("Sending Ack: %v", ack)
			service.Channels.Req <- ack
			service.Cache.Requests = append(service.Cache.Requests, ack)
		case <-service.Channels.Done:
			log.Debug().
				Msg("Received Done signal, shutting down request channel")
			close(service.Channels.Req)
			return
		}
	}
}

func (r *Runner) Stream(service *XDSService) error {
	defer service.Context.cancel()
	defer close(service.Channels.Err)

	var wg sync.WaitGroup
	go func() {
		for {
			wg.Add(1)
			in, err := service.Stream.Recv()
			if err == io.EOF {
				log.Debug().
					Msgf("No more Discovery Responses from %v stream", r.Service.Name)
				close(service.Channels.Res)
				return
			}
			if err != nil {
				log.Err(err).Msgf("error receiving responses on %v stream", r.Service.Name)
				service.Channels.Err <- err
				return
			}
			log.Debug().
				Msgf("Received discovery response: %v", in)
			service.Channels.Res <- in
		}
	}()

	for req := range service.Channels.Req {
		if err := service.Stream.Send(req); err != nil {
			log.Err(err).
				Msg("Error sending discovery request")
			service.Channels.Err <- err
		}
	}
	service.Stream.CloseSend()
	wg.Wait()
	return nil
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


func newAckFromResponse(res *discovery.DiscoveryResponse, initReq *discovery.DiscoveryRequest) *discovery.DiscoveryRequest {
	// Only the first request should need the node ID,
	// so we do not include it in the followups.  If this
	// causes an error, it's a non-conformant error.
	request := &discovery.DiscoveryRequest{
		VersionInfo:   res.VersionInfo,
		ResourceNames: initReq.ResourceNames,
		TypeUrl:       initReq.TypeUrl,
		ResponseNonce: res.Nonce,
	}
	return request
}

func newRequest(resourceNames []string, typeURL, nodeID string) *discovery.DiscoveryRequest {
	return &discovery.DiscoveryRequest{
		VersionInfo: "",
		Node: &core.Node{
			Id: nodeID,
		},
		ResourceNames: resourceNames,
		TypeUrl:       typeURL,
	}
}


