package runner

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/ii/xds-test-harness/internal/db"

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
	Adapter                *ClientConfig
	Target                 *ClientConfig
	NodeID                 string
	Cache                  *Cache
	Aggregated             bool
	Incremental            bool
	Service                *XDSService
	SubscribeRequest       *discovery.DiscoveryRequest
	Delta_SubscribeRequest *discovery.DeltaDiscoveryRequest
	DB                     *db.SQLiteRepository
}

func FreshRunner(current ...*Runner) *Runner {
	var (
		adapter     = &ClientConfig{}
		target      = &ClientConfig{}
		nodeID      = ""
		aggregated  = false
		incremental = false
		DB          = &db.SQLiteRepository{}
	)

	if len(current) > 0 {
		adapter = current[0].Adapter
		target = current[0].Target
		nodeID = current[0].NodeID
		aggregated = current[0].Aggregated
		incremental = current[0].Incremental
		DB = current[0].DB

	}

	return &Runner{
		Adapter:     adapter,
		Target:      target,
		NodeID:      nodeID,
		Cache:       &Cache{},
		Service:     &XDSService{},
		Aggregated:  aggregated,
		Incremental: incremental,
		DB:          DB,
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

func (r *Runner) StartDB() error {
	dbFile := "xds.db"
	os.Remove(dbFile)

	conn, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not open database connection: %v", err))
	}

	DB := db.NewSqliteRepository(conn)
	if err := DB.Migrate(); err != nil {
		return errors.New(fmt.Sprintf("Could not migrate schemas for db: %v", err))
	}

	r.DB = DB
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

func (r *Runner) Delta_Ack(service *XDSService) {
	service.Channels.Delta_Req <- r.Delta_SubscribeRequest
	service.Cache.Delta_Requests = append(service.Cache.Delta_Requests, r.Delta_SubscribeRequest)
	for {
		select {
		case res := <-service.Channels.Delta_Res:
			service.Cache.Delta_Responses = append(service.Cache.Delta_Responses, res)
			ack := delta_newAckFromResponse(res, r.Delta_SubscribeRequest)
			log.Debug().
				Msgf("Sending Ack: %v", ack)
			service.Channels.Delta_Req <- ack
			service.Cache.Delta_Requests = append(service.Cache.Delta_Requests, ack)
		case <-service.Channels.Done:
			log.Debug().
				Msg("Received Done signal, shutting down request channel")
			close(service.Channels.Delta_Req)
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
			if err = r.DB.InsertResponse(in); err != nil {
				service.Channels.Err <- fmt.Errorf("Could not insert response into db: %v", err)
			}
			service.Channels.Res <- in
		}
	}()

	for req := range service.Channels.Req {
		if err := service.Stream.Send(req); err != nil {
			log.Err(err).
				Msg("Error sending discovery request")
			service.Channels.Err <- err
		}
		if err := r.DB.InsertRequest(req); err != nil {
			service.Channels.Err <- fmt.Errorf("Could not insert request into db: %v", err)
		}
	}
	service.Stream.CloseSend()
	wg.Wait()
	return nil
}

func (r *Runner) Delta_Stream(service *XDSService) error {
	defer service.Context.cancel()
	defer close(service.Channels.Err)

	var wg sync.WaitGroup
	go func() {
		for {
			wg.Add(1)
			in, err := service.Delta.Recv()
			if err == io.EOF {
				log.Debug().
					Msgf("No more delta discovery responses from %v delta stream", r.Service.Name)
				close(service.Channels.Delta_Res)
				return
			}
			if err != nil {
				log.Err(err).Msgf("error receiving responses on %v delta stream", r.Service.Name)
				service.Channels.Err <- err
				return
			}
			log.Debug().
				Msgf("Received delta discovery response: %v", in)
			if err := r.DB.InsertResponse(in); err != nil {
				service.Channels.Err <- fmt.Errorf("ya screwed up zach")
			}
			service.Channels.Delta_Res <- in
		}
	}()

	for req := range service.Channels.Delta_Req {
		if err := service.Delta.Send(req); err != nil {
			log.Err(err).
				Msg("Error sending delta discovery request")
			service.Channels.Err <- err
		}
		if err := r.DB.InsertRequest(req); err != nil {
			service.Channels.Err <- fmt.Errorf("[DELTA] Could not insert request into db: %v", err)
		}
	}
	service.Delta.CloseSend()
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

func delta_newAckFromResponse(res *discovery.DeltaDiscoveryResponse, initReq *discovery.DeltaDiscoveryRequest) *discovery.DeltaDiscoveryRequest {
	// Only the first request should need the node ID,
	// so we do not include it in the followups.  If this
	// causes an error, it's a non-conformant error.
	request := &discovery.DeltaDiscoveryRequest{
		TypeUrl:       initReq.TypeUrl,
		ResponseNonce: res.Nonce,
	}
	return request
}

func newRequest(resourceNames []string, typeURL, nodeID string) *discovery.DiscoveryRequest {
	return &discovery.DiscoveryRequest{
		VersionInfo:   "",
		Node:          &core.Node{Id: nodeID},
		ResourceNames: resourceNames,
		TypeUrl:       typeURL,
	}
}

func newDeltaRequest(resourceNames []string, typeURL, nodeID string) *discovery.DeltaDiscoveryRequest {
	return &discovery.DeltaDiscoveryRequest{
		Node:                   &core.Node{Id: nodeID},
		TypeUrl:                typeURL,
		ResourceNamesSubscribe: resourceNames,
	}
}
