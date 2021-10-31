package main

import (
	"context"
	"os"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/ii/xds-test-harness/internal/runner"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

var (
	debug          = pflag.BoolP("debug", "D", false, "sets log level to debug")
	adapterAddress = pflag.StringP("adapter", "A", ":17000", "port of adapter on target")
	targetAddress  = pflag.StringP("target", "T", ":18000", "port of xds target to test")
	nodeID         = pflag.StringP("nodeID", "N", "test-id", "node id of target")
	ADS            = pflag.StringP("ADS", "X", "on", "Whether to include ADS tests, or only run ADS tests. Can be: on, off, or only.")

	godogOpts = godog.Options{
		ShowStepDefinitions: false,
		Randomize:           0,
		StopOnFailure:       false,
		Strict:              false,
		NoColors:            false,
		Format:              "",
		Concurrency:         0,
		Paths:               []string{},
		Output:              colors.Colored(os.Stdout),
	}

	r *runner.Runner
)

func init() {
	godog.BindCommandLineFlags("godog.", &godogOpts)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {
		r = runner.FreshRunner()
		if err := r.ConnectClient("target", *targetAddress); err != nil {
			log.Fatal().
				Msgf("error connecting to target: %v", err)
		}
		if err := r.ConnectClient("adapter", *adapterAddress); err != nil {
			log.Fatal().
				Msgf("error connecting to adapter: %v", err)
		}
		r.NodeID = *nodeID
		log.Info().
			Msgf("Connected to target at %s and adapter at %s\n", *targetAddress, *adapterAddress)
	})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		log.Debug().Msg("Fresh Runner!")
		r = runner.FreshRunner(r)
		return ctx, nil
	})
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		c := pb.NewAdapterClient(r.Adapter.Conn)
		clearRequest := &pb.ClearRequest{
			Node: r.NodeID,
		}
		clear, err := c.ClearState(context.Background(), clearRequest)
		if err != nil {
			log.Err(err).
				Msg("Couldn't clear state")
		}
		log.Debug().
			Msgf("Clearing state...%v", clear.Response)
		return ctx, nil
	})
	r.LoadSteps(ctx)
}

func main() {
	pflag.Parse()
	godogOpts.Paths = pflag.Args()
	if *ADS == "off" {
		godogOpts.Tags = "~@ADS"
	}
	if *ADS == "only" {
		godogOpts.Tags = "@ADS"
	}
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	status := godog.TestSuite{
		Name:                 "xDS Test Suite",
		ScenarioInitializer:  InitializeScenario,
		TestSuiteInitializer: InitializeTestSuite,
		Options:              &godogOpts,
	}.Run()
	os.Exit(status)
}
