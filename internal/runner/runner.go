package runner

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ii/xds-test-harness/internal/parser"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	any "google.golang.org/protobuf/types/known/anypb"
)

var (
	opts []grpc.DialOption = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}
)

type ClientConfig struct {
	Port string
	Conn *grpc.ClientConn
}

type Cache struct {
	// StartState     *pb.Snapshot
	// StateSnapshots []*pb.Snapshot
	FinalResponse *discovery.DiscoveryResponse
}

type ValidateResource struct {
	Version string
	Nonce   string
}

type Validate struct {
	RequestCount     int
	ResponseCount    int
	Resources        map[string]map[string]ValidateResource
	RemovedResources map[string]map[string]ValidateResource
}

func NewValidate() *Validate {
	resources := make(map[string]map[string]ValidateResource)
	removed := make(map[string]map[string]ValidateResource)
	return &Validate{
		RequestCount:     0,
		ResponseCount:    0,
		Resources:        resources,
		RemovedResources: removed,
	}
}

type Runner struct {
	Adapter          *ClientConfig
	Target           *ClientConfig
	NodeID           string
	Cache            *Cache
	Aggregated       bool
	Incremental      bool
	Service          *XDSService
	SubscribeRequest *any.Any
	Validate         *Validate
}

func FreshRunner(current ...*Runner) *Runner {
	var (
		adapter     = &ClientConfig{}
		target      = &ClientConfig{}
		nodeID      = ""
		aggregated  = false
		incremental = false
	)

	if len(current) > 0 {
		adapter = current[0].Adapter
		target = current[0].Target
		nodeID = current[0].NodeID
		aggregated = current[0].Aggregated
		incremental = current[0].Incremental

	}

	validate := NewValidate()

	return &Runner{
		Adapter:     adapter,
		Target:      target,
		NodeID:      nodeID,
		Cache:       &Cache{},
		Service:     &XDSService{},
		Aggregated:  aggregated,
		Incremental: incremental,
		Validate:    validate,
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
	for {
		select {
		case res := <-service.Channels.Res:
			ack, _ := r.newAckFromResponse(res)
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

func (r *Runner) SotwStream(service *XDSService) {
	sotw := service.Sotw
	ch := service.Channels
	defer sotw.Context.cancel()
	defer close(service.Channels.Err)

	// Our Response loop
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
				log.Debug().Err(err)
				ch.Err <- err
				return
			}
			log.Debug().
				Msgf("Received discovery response: %v", in)

			resources, err := parser.ResourceNames(in)
			if err != nil {
				ch.Err <- fmt.Errorf("could not gather resource names from response: %v", err)
				return
			}
			log.Debug().Msgf("Verison: %v", in.VersionInfo)
			for _, resource := range resources {
				r.Validate.Resources[in.TypeUrl][resource] = ValidateResource{
					Version: in.VersionInfo,
					Nonce:   in.Nonce,
				}
			}
			r.Validate.ResponseCount++
			res, err := any.New(in)
			if err != nil {
				ch.Err <- err
			}
			ch.Res <- res
		}
	}()

	// Our requests loop
	for req := range service.Channels.Req {
		var dr discovery.DiscoveryRequest
		if err := req.UnmarshalTo(&dr); err != nil {
			service.Channels.Err <- fmt.Errorf("error unmarshalling from request channel: %v", err)
			return
		}
		if err := sotw.Stream.Send(&dr); err != nil {
			log.Debug().Msgf("error sending: %v", err)
			service.Channels.Err <- fmt.Errorf("error sending discovery request: %v", err)
		}
		r.Validate.RequestCount++
	}
	if err := sotw.Stream.CloseSend(); err != nil {
		ch.Err <- err
	}
	wg.Wait()
}

// Bidirectioinal stream between client and server. Listens for any
// responses from server and sends them along the response channel.
// Listens to new requests from the request channel and sends them along to the server.
func (r *Runner) DeltaStream(service *XDSService) {
	delta := service.Delta
	ch := service.Channels
	defer delta.Context.cancel()
	defer close(service.Channels.Err)

	// Our response loop
	var wg sync.WaitGroup
	go func() {
		for {
			wg.Add(1)
			in, err := delta.Stream.Recv()
			if err == io.EOF {
				log.Debug().
					Msgf("[Delta] No more Discovery Responses from %v stream", r.Service.Name)
				close(ch.Res)
				return
			}
			if err != nil {
				ch.Err <- fmt.Errorf("[Delta] Error receiving discovery response: %v", err)
				return
			}
			log.Debug().
				Msgf("[Delta] Received discovery response: %v", in)
			for _, resource := range in.GetResources() {
				r.Validate.Resources[in.TypeUrl][resource.Name] = ValidateResource{
					Version: in.SystemVersionInfo,
					Nonce:   in.Nonce,
				}
				delete(r.Validate.RemovedResources[in.TypeUrl], resource.Name)
			}
			for _, removed := range in.GetRemovedResources() {
				r.Validate.RemovedResources[in.TypeUrl][removed] = ValidateResource{
					Nonce: in.Nonce,
				}
			}
			r.Validate.ResponseCount++
			res, err := any.New(in)
			if err != nil {
				ch.Err <- err
			}
			ch.Res <- res
		}
	}()

	// Our request loop
	for req := range service.Channels.Req {
		var request discovery.DeltaDiscoveryRequest
		if err := req.UnmarshalTo(&request); err != nil {
			service.Channels.Err <- fmt.Errorf("[Delta] Error unmarshalling request from anypb message: %v", err)
		}
		if err := delta.Stream.Send(&request); err != nil {
			service.Channels.Err <- fmt.Errorf("[Delta] Error sending discovery request: %v", err)
		}
		r.Validate.RequestCount++
	}
	if err := delta.Stream.CloseSend(); err != nil {
		ch.Err <- err
	}
	wg.Wait()
}

func (r *Runner) Stream(service *XDSService) {
	if r.Incremental {
		r.DeltaStream(service)
	} else {
		r.SotwStream(service)
	}
}

func connectViaGRPC(client *ClientConfig, server string) (conn *grpc.ClientConn, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err = grpc.DialContext(ctx, client.Port, opts...)
	cancel()
	if err != nil {
		err = fmt.Errorf("cannot connect at %v: %v", client.Port, err)
		return nil, err
	}
	log.Debug().
		Msgf("Runner connected to %v", server)
	return conn, nil
}

// Using the last response and current subscribing request, create a new DiscoveryRequest to ACK that response.
// We use the current subscribing request for the cases where the client is subscribing to A,B,C but only A,B
// exist.  In that case, we want to ack that we've received A,B but that we are STILL subscribing to A,B,C.
func (r *Runner) newAckFromResponse(res *any.Any) (*any.Any, error) {
	// Only the first request should need the node ID,
	// so we do not include it in the followups.  If this
	// causes an error, it's a non-conformant error.
	if r.Incremental {
		var response discovery.DeltaDiscoveryResponse
		if err := res.UnmarshalTo(&response); err != nil {
			return nil, err
		}
		request := &discovery.DeltaDiscoveryRequest{
			TypeUrl:       response.TypeUrl,
			ResponseNonce: response.Nonce,
		}
		ack, err := any.New(request)
		return ack, err
	} else {
		var sub discovery.DiscoveryRequest
		var response discovery.DiscoveryResponse
		if err := r.SubscribeRequest.UnmarshalTo(&sub); err != nil {
			return nil, err
		}
		if err := res.UnmarshalTo(&response); err != nil {
			return nil, err
		}
		request := &discovery.DiscoveryRequest{
			VersionInfo:   response.VersionInfo,
			ResourceNames: sub.ResourceNames,
			TypeUrl:       sub.TypeUrl,
			ResponseNonce: response.Nonce,
		}
		ack, err := any.New(request)
		if err != nil {
			return nil, err
		}
		return ack, err
	}
}

func (r *Runner) newRequest(resourceNames []string, typeURL string) *any.Any {
	if r.Incremental {
		request := &discovery.DeltaDiscoveryRequest{
			Node:                   &core.Node{Id: r.NodeID},
			TypeUrl:                typeURL,
			ResourceNamesSubscribe: resourceNames,
		}
		any, _ := any.New(request)
		return any
	} else {
		request := &discovery.DiscoveryRequest{
			VersionInfo: "",
			Node: &core.Node{
				Id: r.NodeID,
			},
			ResourceNames: resourceNames,
			TypeUrl:       typeURL,
		}
		any, _ := any.New(request)
		return any
	}
}
