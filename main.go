package main

import (
	"context"
	"fmt"
	// "fmt"
	"os"

	"strings"

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
	variants       = pflag.StringP("variants", "V", "1111", "Set which xDS transport variants your server supports.\n These are, in order: sotw non-aggregated\n, sotw aggregated\n, incremental non-aggregated\n, incremental aggregated\n. 1 if your server supports that variant, 0 if not\n. eg: --variants 1010 supports sotw non-aggregated and incremental non-aggregated.")
	aggregated     = false
	incremental    = false

	godogOpts = godog.Options{
		ShowStepDefinitions: false,
		Randomize:           0,
		StopOnFailure:       false,
		Strict:              false,
		NoColors:            false,
		Tags:                "",
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
		r.Aggregated = aggregated
		log.Info().
			Msgf("Connected to target at %s and adapter at %s\n", *targetAddress, *adapterAddress)
		if r.Aggregated {
			log.Info().
				Msgf("Tests will be run via ADS")
		} else {
			log.Info().
				Msgf("Tests will be run non-aggregated, via separate streams")
		}
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

func supportedVariants(combo string) (err error, supportedVariants []bool) {
	flags := strings.Split(combo, "")
	if len(flags) < 4 {
		err := fmt.Errorf("Expected four digits, each a 1 or 0 (e.g. 1001). Given \"%v\"", combo)
		return err, []bool{}
	}
	for i := 0; i < 4; i++ {
		if flags[i] == "1" {
			supportedVariants = append(supportedVariants, true)
		} else if flags[i] == "0" {
			supportedVariants = append(supportedVariants, false)
		} else {
			err := fmt.Errorf("Expected four digits, each a 1 or 0 (e.g. 1001). Given \"%v\"", combo)
			return err, []bool{}
		}
	}
	return nil, supportedVariants
}

func main() {
	pflag.Parse()
	godogOpts.Paths = pflag.Args()
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	suite := godog.TestSuite{
		Name:                 "xDS Test Suite",
		ScenarioInitializer:  InitializeScenario,
		TestSuiteInitializer: InitializeTestSuite,
		Options:              &godogOpts,
	}

	// we have four variants, either set to T or F
	err, supportedVariants := supportedVariants(*variants)
	if err != nil {
		log.Info().Msgf("Error parsing variants config: %v\n", err)
		os.Exit(0)
	}

	//SOTW, Separate
	if supportedVariants[0] {
		incremental = false
		aggregated = false
		godogOpts.Tags = strings.Join([]string{godogOpts.Tags, "@sotw","@separate"}, " && ")
		suite.Run();
	}

	//SOTW, Aggregated
	if supportedVariants[1] {
		incremental = false
		aggregated = true
		godogOpts.Tags = strings.Join([]string{godogOpts.Tags, "@sotw","@aggregated"}, " && ")
		suite.Run();
	}

	//Incremental, Separate
	if supportedVariants[2] {
		incremental = true
		aggregated = false
		godogOpts.Tags = strings.Join([]string{godogOpts.Tags, "@incremental","@separate"}, " && ")
		suite.Run();
	}

	//Incremental, Aggregated
	if supportedVariants[3] {
		incremental = true
		aggregated = true
		godogOpts.Tags = strings.Join([]string{godogOpts.Tags, "@incremental","@aggregated"}, " && ")
		suite.Run();
	}
	os.Exit(0)
}
