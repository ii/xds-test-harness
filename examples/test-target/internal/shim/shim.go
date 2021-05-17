package shim

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/zachmandeville/tester-prototype/examples/test-target/internal/processor"
	pb "github.com/zachmandeville/tester-prototype/api/shim"
)

type shimServer struct {
	pb.UnimplementedShimServer
}

var (
	proc *processor.Processor
)

func (s *shimServer) GiveCompliment(ctx context.Context, req *pb.ComplimentRequest) (res *pb.ComplimentResponse, err error) {
	name := req.Name
	compliment := &pb.ComplimentResponse{
		Compliment: fmt.Sprintf("Hi, %v, you are GREAT!", name),
	}
	return compliment, nil
}

func (s *shimServer) AddCluster(ctx context.Context, req *pb.ClusterRequest) (res *pb.ClusterResponse, err error) {
	cluster := req.Cluster
	snapshot, err := proc.UpdateSnapshot(cluster)
	if err != nil {
		log.Printf("error updating snapshot: %v", err)
	}
	snapJSON, err:= json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		fmt.Printf("error marshalling snapshot to json:%v", err)
	}
	fmt.Printf("snapshot JSON: %v", string(snapJSON))
	response := &pb.ClusterResponse{
		Message: fmt.Sprintf("A cluster named %v was added\n", cluster),
		Snapshot: string(snapJSON),
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
	pb.RegisterShimServer(s, &shimServer{})
	log.Printf("shim listening on %v\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to server: %v", err)
	}
}
