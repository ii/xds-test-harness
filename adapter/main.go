package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/api/adapter"
	shimpb "github.com/zachmandeville/tester-prototype/api/shim"
)

const (
	port     = ":6767"
	shimPort = "localhost:17000"
)

type server struct {
	pb.UnimplementedAdapterServer
}

type Target struct {
	Port       string
	Connection *grpc.ClientConn
}

type Shim struct {
	Port string
	Connection *grpc.ClientConn
}

var (
	target *Target = nil
	shim *Shim = nil
)

func (s *server) GiveCompliments(name *pb.Name, stream pb.Adapter_GiveComplimentsServer) error {
	adjectives := []string{"cool", "fun", "smart", "awesome"}
	for i := 0; i <= 28; i++ {
		adjective := adjectives[i%len(adjectives)]
		compliment := fmt.Sprintf("You, %v, are %v", name.Name, adjective)
		if err := stream.Send(&pb.Compliment{Message: compliment}); err != nil {
			log.Fatalf("could not send compliment: %v", err)
		}
	}
	return nil
}

func (s *server) ConnectToShim(ctx context.Context, req *pb.ShimRequest) (res *pb.ShimResponse, err error) {
	port := req.Port
	conn, err := grpc.Dial(port, grpc.WithInsecure(), grpc.WithTimeout(time.Second *5), grpc.WithBlock())
	if err != nil {
		log.Printf("error connecting to shim: %v", err)
		return nil, err
	}
	shim = &Shim{
		Port:       port,
		Connection: conn,
	}
	response := &pb.ShimResponse{
		Message: "connected to shim",
	}
	return response, nil
}

func (s *server) RegisterResource(ctx context.Context, in *pb.ResourceSpec) (*pb.Snapshot, error) {
	c := shimpb.NewShimClient(shim.Connection)
	clusters := in.Spec
	req := &shimpb.ClusterRequest{
		Cluster: clusters,
	}
	snapshot, err := c.AddCluster(context.Background(), req)
	if err != nil {
		log.Printf("failed to add a cluster: %v\n", err)
	}
	response := &pb.Snapshot{
		Snapshot: snapshot.Snapshot,
	}
	return response, nil
}

func main() {
	go func() {
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterAdapterServer(s, &server{})
		fmt.Printf("Adapter Server started on port %v\n", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	for {
		if 1 == 2 {
			fmt.Println("hi")
		}
	}
}
