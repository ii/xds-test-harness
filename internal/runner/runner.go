package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"io"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	cds "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	lds "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"github.com/ii/xds-test-harness/internal/parser"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"google.golang.org/grpc"
	"github.com/rs/zerolog/log"
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
	StartState *pb.Snapshot
	StateSnapshots []*pb.Snapshot
	FinalResponse *discovery.DiscoveryResponse
}

type Service struct {
	Req   chan *discovery.DiscoveryRequest
	Res   chan *discovery.DiscoveryResponse
	Err   chan error
	Done  chan bool
	Cache struct {
		InitResource []string
		Requests  []*discovery.DiscoveryRequest
		Responses []*discovery.DiscoveryResponse
	}
}

type Runner struct {
	Adapter *ClientConfig
	Target  *ClientConfig
	NodeID  string
	Cache   *Cache
	CDS     *Service
	LDS     *Service
}

func FreshRunner (current ...*Runner) *Runner {
	var (
	 adapter = &ClientConfig{}
	 target = &ClientConfig{}
	 nodeID = ""
	)

	if len(current) > 0 {
		adapter = current[0].Adapter
		target = current[0].Target
		nodeID = current[0].NodeID

	}

	return &Runner{
		Adapter: adapter,
		Target: target,
		NodeID: nodeID,
		Cache: &Cache{},
		CDS: &Service{},
		LDS: &Service{},
	}
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

func (r *Runner) ConnectClient(server, address string) error {
	var client *ClientConfig
	if server == "target" {
		client = r.Target
	}
	if server == "adapter" {
		client = r.Adapter
	}
	client.Port = address
	conn, err := connectViaGRPC(client, server)
	if err != nil {
		return err
	}
	client.Conn = conn
	return nil
}

func (r *Runner) NewRequest(resourceList []string, typeURL string) *discovery.DiscoveryRequest {
	resourceNames := []string{}
	for _, name := range resourceList {
		resourceNames = append(resourceNames, name)
	}
	return &discovery.DiscoveryRequest{
		VersionInfo: "",
		Node: &core.Node{
			Id: r.NodeID,
		},
		ResourceNames: resourceNames,
		TypeUrl:       typeURL,
	}
}

func (r *Runner) NewAckFromResponse(res *discovery.DiscoveryResponse, initReq *discovery.DiscoveryRequest) (*discovery.DiscoveryRequest, error) {
	response, err := parser.ParseDiscoveryResponseV2(res)
	if err != nil {
		err := fmt.Errorf("error parsing dres for acking: %v", err)
		return nil, err
	}

	// Only the first request should need the node ID,
	// so we do not include it in the followups.  If this
	// causes an error, it's a non-conformant error.
	request := &discovery.DiscoveryRequest{
		VersionInfo:   response.Version,
		ResourceNames: initReq.ResourceNames,
		TypeUrl:       initReq.TypeUrl,
		ResponseNonce: response.Nonce,
	}

	return request, nil
}

func (r *Runner) Ack (initReq *discovery.DiscoveryRequest, service *Service) {
	service.Req <- initReq
	service.Cache.Requests = append(service.Cache.Requests, initReq)
	for {
		select {
		case res := <- service.Res:
			service.Cache.Responses = append(service.Cache.Responses, res)
			ack, err := r.NewAckFromResponse(res, initReq)
			if err != nil {
				service.Err <- err
				return
			}
			log.Debug().
				Msgf("Sending Ack: %v", ack)
			service.Req <- ack
	        service.Cache.Requests = append(service.Cache.Requests, ack)
		case <- service.Done:
			log.Debug().Msg("Received Done signal, shutting down request channel")
			close(service.Req)
			return
		}
	}
}

func (r *Runner) Zack (initReq *discovery.DiscoveryRequest, service *xDSService) {
	service.Channels.Req <- initReq
	service.Cache.Requests = append(service.Cache.Requests, initReq)
	for {
		select {
		case res := <- service.Channels.Res:
			service.Cache.Responses = append(service.Cache.Responses, res)
			ack, err := r.NewAckFromResponse(res, initReq)
			if err != nil {
				service.Channels.Err <- err
				return
			}
			log.Debug().
				Msgf("Sending Ack: %v", ack)
			service.Channels.Req <- ack
	        service.Cache.Requests = append(service.Cache.Requests, ack)
		case <- service.Channels.Done:
			log.Debug().Msg("Received Done signal, shutting down request channel")
			close(service.Channels.Req)
			return
		}
	}
}

func (r *Runner) LDSStream() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer close(r.LDS.Err)

		client := lds.NewListenerDiscoveryServiceClient(r.Target.Conn)
		stream, err := client.StreamListeners(ctx)
		if err != nil {
			err = errors.New(fmt.Sprintf("Cannot start LDS stream %v. error: %v", stream, err))
			log.Debug().
				Err(err).
				Msg("")
			r.LDS.Err <- err
		}
		var wg sync.WaitGroup
		go func() {
			for {
				wg.Add(1)
				in, err := stream.Recv()
				if err == io.EOF {
					log.Debug().
						Msg("No more Discovery Responses from LDS stream")
					close(r.LDS.Res)
					return
				}
				if err != nil {
					log.Err(err).Msg("error receiving responses on LDS stream")
					r.LDS.Err <- err
					return
				}
				log.Debug().
					Msgf("Received discovery response: %v", in)
				r.LDS.Res <- in
			}
		}()

	for req := range r.LDS.Req {
		if err := stream.Send(req); err != nil {
			log.Err(err).
				Msg("Error sending discovery request")
			r.LDS.Err <- err
		}
	}
	stream.CloseSend()
	wg.Wait()
	return err
}

func (r *Runner) Stream(service *xDSService) error {
	defer close(service.Channels.Err)

	var wg sync.WaitGroup
	go func() {
		for {
			wg.Add(1)
			in, err := service.Stream.Recv()
			if err == io.EOF {
				log.Debug().
					Msg("No more Discovery Responses from LDS stream")
				close(service.Channels.Res)
				return
			}
			if err != nil {
				log.Err(err).Msg("error receiving responses on LDS stream")
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


func (r *Runner) CDSStream() error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer close(r.CDS.Err)

	client := cds.NewClusterDiscoveryServiceClient(r.Target.Conn)
	stream, err := client.StreamClusters(ctx)
	if err != nil {
		err = errors.New(fmt.Sprintf("Cannot start CDS stream %v. error: %v", stream, err))
		log.Error().
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
					Msg("No more Discovery Responses from CDS stream")
				close(r.CDS.Res)
				return
			}
			if err != nil {
				log.Err(err).Msg("error receiving responses on CDS stream")
				r.CDS.Err <- err
				return
			}
			log.Debug().
				Msgf("Received discovery response: %v", in)
			r.CDS.Res <- in
		}
	}()

	for req := range r.CDS.Req {
		if err := stream.Send(req); err != nil {
			log.Err(err).
				Msg("Error sending discovery request")
			r.CDS.Err <- err
		}
	}
	stream.CloseSend()
	wg.Wait()
	return err
}
