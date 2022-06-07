package runner

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/ii/xds-test-harness/internal/parser"

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

type ValidateResource struct {
	Version string
	Nonce   string
}

type Validate struct {
	RequestCount  int
	ResponseCount int
	Resources     map[string]map[string]ValidateResource
}

func NewValidate() *Validate {
	resources := make(map[string]map[string]ValidateResource)
	return &Validate{
		RequestCount:  0,
		ResponseCount: 0,
		Resources:     resources,
	}
}

type Runner struct {
	Adapter          *ClientConfig
	Target           *ClientConfig
	NodeID           string
	Cache            *Cache
	Aggregated       bool
	Service          *XDSService
	SubscribeRequest *discovery.DiscoveryRequest
	Validate         *Validate
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

	validate := NewValidate()

	return &Runner{
		Adapter:    adapter,
		Target:     target,
		NodeID:     nodeID,
		Cache:      &Cache{},
		Service:    &XDSService{},
		Aggregated: aggregated,
		Validate:   validate,
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
	// service.Cache.Requests = append(service.Cache.Requests, r.SubscribeRequest)
	for {
		select {
		case res := <-service.Channels.Res:
			dr := res.(*discovery.DiscoveryResponse)
			ack := newAckFromResponse(dr, r.SubscribeRequest)
			log.Debug().
				Msgf("Sending Ack: %v", ack)
			service.Channels.Req <- ack
		case <-service.Channels.Done:
			log.Debug().
				Msg("Received Done signal, shutting down request channel")
			close(service.Channels.Req)
			return
		}
	}
}

func (r *Runner) Stream(service *XDSService) error {
	sotw := service.Sotw
	ch := service.Channels
	defer sotw.Context.cancel()
	defer close(service.Channels.Err)

	var wg sync.WaitGroup
	go func() {
		for {
			wg.Add(1)
			in, err := sotw.Stream.Recv()
			if err == io.EOF {
				log.Debug().
					Msgf("No more Discovery Responses from %v stream", r.Service.Name)
				close(ch.Res)
				return
			}
			if err != nil {
				ch.Err <- err
				return
			}
			log.Debug().
				Msgf("Received discovery response: %v", in)

			resources, err := parser.ResourceNames(in)
			if err != nil {
				log.Err(err).Msg("Could not gather resource names from response")
				ch.Err <- err
				return
			}
			for _, resource := range resources {
				r.Validate.Resources[in.TypeUrl][resource] = ValidateResource{
					Version: in.VersionInfo,
					Nonce:   in.Nonce,
				}
			}
			r.Validate.ResponseCount++
			ch.Res <- in
		}
	}()

	for req := range service.Channels.Req {
		dr := req.(*discovery.DiscoveryRequest)
		if err := sotw.Stream.Send(dr); err != nil {
			log.Err(err).
				Msg("Error sending discovery request")
			service.Channels.Err <- err
		}
		r.Validate.RequestCount++
	}
	sotw.Stream.CloseSend()
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

// Using the last response and current subscribing request, create a new DiscoveryRequest to ACK that response.
// We use the current subscribing request for the cases where the client is subscribing to A,B,C but only A,B
// exist.  In that case, we want to ack that we've received A,B but that we are STILL subscribing to A,B,C.
func newAckFromResponse(res *discovery.DiscoveryResponse, subscribingReq *discovery.DiscoveryRequest) *discovery.DiscoveryRequest {
	// Only the first request should need the node ID,
	// so we do not include it in the followups.  If this
	// causes an error, it's a non-conformant error.
	request := &discovery.DiscoveryRequest{
		VersionInfo:   res.VersionInfo,
		ResourceNames: subscribingReq.ResourceNames,
		TypeUrl:       subscribingReq.TypeUrl,
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
