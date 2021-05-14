package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/adapter/adapter"
	shim "github.com/zachmandeville/tester-prototype/test-target/test-target"
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
	conn, err := grpc.Dial(req.Port, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
	if err != nil {
		fmt.Printf("Error dialing into %v: %v", req.Port, err)
		return &pb.ConnectionResponse{}, err
	}
	target = &Target{
		Port:       req.Port,
		Connection: conn,
	}
	response := &pb.ConnectionResponse{
		Message: "Connected to test target.",
	}
	return response, nil
}

func getCompliment(c shim.ShimClient, name string) {
	request := &shim.ComplimentRequest{
		Name: name,
	}
	compliment, err := c.GiveCompliment(context.Background(), request)
	if err != nil {
		log.Printf("failed to get compliment: %v\n", err)
	}
	fmt.Println(compliment.Compliment)
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

	go func() {
	fmt.Println("\nConnecting to Shim")
	conn, err := grpc.Dial(shimPort, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
	if err != nil {
		log.Fatalf("unable to connect to %v: %v", shimPort, err)
	}
	defer conn.Close()
	c := shim.NewShimClient(conn)
	getCompliment(c, "zach")
	}()
	for {
		if 1 == 2 {
			fmt.Println("hi")
		}
	}
}
