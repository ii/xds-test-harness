package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/example/complimenter"
)

const (
	address = ":6767"
)

type runner struct {
	compliment string
	conn       *grpc.ClientConn
}

func (r *runner) aConnectionToTheComplimenter() error {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second * 5))
	if err != nil {
		fmt.Printf("error connecting to adapter: %v", err)
		return err
	}
	r.conn = conn
	return nil
}

func (r *runner) iReceiveACompliment() error {
	compliment := r.compliment
	if compliment == "" {
		return fmt.Errorf("no compliment recieve: %v", r.compliment)
	}
	return nil
}

func (r *runner) iSendItARequestWithMyName() error {
	c := pb.NewComplimenterClient(r.conn)
	request := &pb.ComplimentRequest{
		Name: "repo reader",
	}
	success, err := c.GiveCompliment(context.Background(), request)
	if err != nil {
		fmt.Printf("error in requesting compliment: %v", err)
		return err
	}
	r.compliment = success.Compliment
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	r := &runner{}
	ctx.Step(`^a connection to the complimenter$`, r.aConnectionToTheComplimenter)
	ctx.Step(`^I receive a compliment$`, r.iReceiveACompliment)
	ctx.Step(`^I send it a request with my name$`, r.iSendItARequestWithMyName)
}
