package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/examples/grpc-godog-test/complimenter"
)

const (
	port = ":6767"
)

type server struct {
	pb.UnimplementedComplimenterServer
}

func (s *server) GiveCompliment(ctx context.Context, req *pb.ComplimentRequest) (res *pb.ComplimentResponse, err error) {
	name := req.Name
	compliment := fmt.Sprintf("You, %v, are awesome!", name)

	response := &pb.ComplimentResponse{
		Compliment: compliment,
	}
	return response, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterComplimenterServer(s, &server{})
	fmt.Printf("Compliment Server started on port %v", port)
	if err := s.Serve(lis); err != nil {log.Fatalf("Failed to server: %v", err)
	}
}
