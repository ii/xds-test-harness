package adapter

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/api/adapter"
	"github.com/zachmandeville/tester-prototype/examples/test-target/internal/processor"
)

type adapterServer struct {
	pb.UnimplementedAdapterServer
}

var (
	proc *processor.Processor
)

func (a *adapterServer) SetState(ctx context.Context, in *pb.Snapshot) (response *pb.SetStateResponse, err error) {
	_, err = proc.UpdateSnapshot(in)
	if err != nil {
		fmt.Printf("error updating snapshot: %v", err)
	}
	response = &pb.SetStateResponse{
		Message: "Success",
	}
	return response, nil
}

func RunServer(p *processor.Processor, port string) {
	proc = p
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterAdapterServer(s, &adapterServer{})
	log.Printf("Adapter listening on %v\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Adapter failed to serve: %v", err)
	}
}
