package main

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"google.golang.org/grpc"

	pb "github.com/zachmandeville/tester-prototype/adapter/adapter"
)

type Results struct {
	target string
	message string
}

func NewResults () *Results {
  return &Results {
	  target: "",
	  message: "",
  }
}

type runner struct {
	results *Results
	adapter *grpc.ClientConn
}

func (r *runner) aTargetAddress() error {
	r.results.target = "localhost:18000"
	return nil
}

func (r *runner) iAttemptToConnectToTheAddress() error {
	conn, err := grpc.Dial("localhost:6767", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("error connecting to adapter: %v", err)
	}
	r.adapter = conn
	c := pb.NewAdapterClient(conn)
	target := &pb.ConnectionRequest{
		Port: r.results.target,
	}
	success, err := c.ConnectToTarget(context.Background(), target)
	if err != nil {
		fmt.Printf("errrrrrrrr....%v\n", err)
	}
	r.results.message = success.Message
	return nil
}

func (r *runner) iGetASuccessMessage() error {
	if r.results.message == "Connected to test target." {
		return nil
	} else {
		return godog.ErrPending
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	runner := &runner{}
	ctx.BeforeScenario(func (sc *godog.Scenario) {
		runner.results = NewResults();
	})
	ctx.Step(`^a target address$`, runner.aTargetAddress)
	ctx.Step(`^I attempt to connect to the address$`, runner.iAttemptToConnectToTheAddress)
	ctx.Step(`^I get a success message$`, runner.iGetASuccessMessage)
}
