package main

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/ii/xds-test-harness/internal/runner"
	"github.com/spf13/pflag"
	"os"
)

var godogOpts = godog.Options{Output: colors.Colored(os.Stdout)}
var r *runner.Runner
var adapterAddress = pflag.StringP("adapter", "A", ":17000", "port of adapter on target")
var targetAddress = pflag.StringP("target", "T", ":18000", "port of xds target to test")
var nodeID = pflag.StringP("nodeID", "N", "test-id", "node id of target")

func init() {
	godog.BindCommandLineFlags("godog.", &godogOpts)
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {
		r = runner.NewRunner()
		if err := r.ConnectToTarget(*targetAddress); err != nil {
			fmt.Printf("error connecting to target: %v", err)
			os.Exit(1)
		}
		if err := r.ConnectToAdapter(*adapterAddress); err != nil {
			fmt.Printf("error connecting to adapter: %v", err)
			os.Exit(1)
		}
		r.NodeID = *nodeID
		fmt.Printf("Connected to target at %s and adapter at %s\n", *targetAddress, *adapterAddress)
	})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	r.LoadSteps(ctx)
}

func main() {
	pflag.Parse()
	godogOpts.Paths = pflag.Args()

	status := godog.TestSuite{
		Name:                 "xDS Test Suite",
		ScenarioInitializer:  InitializeScenario,
		TestSuiteInitializer: InitializeTestSuite,
		Options:              &godogOpts,
	}.Run()
	os.Exit(status)
}
