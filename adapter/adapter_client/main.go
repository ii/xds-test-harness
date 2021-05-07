package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/adapter/adapter"
)

const (
	adapterAddress = "localhost:6767"
	targetAddress = "localhost:18000"
)

func streamCompliments (c pb.AdapterClient , name *pb.Name) {
	stream, err := c.GiveCompliments(context.Background(), name)
	if err != nil {
		log.Fatalf("unable to start receiving compliments: %v", err)
	}
	for {
		compliment, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Compliment messed up in some way: %v", err)
		}
		log.Println(compliment)
	}
}

func connectToTarget (c pb.AdapterClient, address string) {
	target := &pb.ConnectionRequest{
		Port: address,
	}
	state, err := c.ConnectToTarget(context.Background(), target)
	if err != nil {
		log.Fatalf("errrrrrrrr....%v", err)
	}
	fmt.Printf("Target state: %v", state)
}

func main () {
	fmt.Println("Client Started")
	conn, err := grpc.Dial(adapterAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("unable to connect to %v: %v", adapterAddress, err)
	}
	defer conn.Close()

	c := pb.NewAdapterClient(conn)

	name := &pb.Name{
		Name: "Caleb",
	}
	streamCompliments(c, name)
	connectToTarget(c,targetAddress)
}
