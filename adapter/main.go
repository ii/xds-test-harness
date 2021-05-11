package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/adapter/adapter"
)

const (
	port = ":6767"
)

type server struct {
	pb.UnimplementedAdapterServer
}

type Target struct {
    Port       string
    Connection *grpc.ClientConn
}

var (
    target *Target = nil
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

func (s *server) ConnectToTarget(ctx context.Context, req *pb.ConnectionRequest) (res *pb.ConnectionResponse, err error) {
    fmt.Printf("Connecting to test target at %v\n", req.Port)
    conn, err := grpc.Dial(req.Port, grpc.WithInsecure(), grpc.WithBlock(),grpc.WithTimeout(time.Second * 5))
    if err != nil {
        fmt.Printf("Error dialing into %v: %v", req.Port, err)
		  return &pb.ConnectionResponse{}, err
    }
    target = &Target{
        Port: req.Port,
        Connection: conn,
    }
    response := &pb.ConnectionResponse{
        Message: "Connected to test target.",
    }
    return response, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterAdapterServer(s, &server{})
	fmt.Printf("Compliment Server started on port %v", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to server: %v", err)
	}
}
