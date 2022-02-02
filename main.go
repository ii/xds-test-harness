package main

import (
	"context"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/ii/xds-test-harness/internal/runner"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

var (
	debug          = pflag.BoolP("debug", "D", false, "sets log level to debug")
	config         = pflag.StringP("config", "C", "", "Path to optional config file. This file sets the adapter and target addresses and supported variants.")
	adapterAddress = pflag.StringP("adapter", "A", ":17000", "port of adapter on target")
	targetAddress  = pflag.StringP("target", "T", ":18000", "port of xds target to test")
	nodeID         = pflag.StringP("nodeID", "N", "test-id", "node id of target")
	variant        = pflag.StringArrayP("variant", "V", []string{"sotw non-aggregated", "sotw aggregated", "incremental non-aggregated", "incremental aggregated"}, "xDS protocol variant your server supports. Add a separate flag per each supported variant.\n Possibleariants are: sotw non-aggregated\n, sotw aggregated\n, incremental non-aggregated\n, incremental aggregated\n.")
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
			Msgf("Connected to target at port %s and adapter at port %s\n", *targetAddress, *adapterAddress)
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

func parseSupportedVariants(variants []string) (err error, supported map[string]bool) {
	supported = make(map[string]bool)
	for _, variant := range variants {
		variant = strings.ToLower(strings.TrimSpace(variant))
		switch variant {
		case "sotw non-aggregated":
			supported["sotw non-aggregated"] = true
		case "sotw aggregated":
			supported["sotw aggregated"] = true
		case "incremental non-aggregated":
			supported["incremental non-aggregated"] = true
		case "incremental aggregated":
			supported["incremental aggregated"] = true
		default:
			log.Info().Msgf("Cannot recognize variant: %v\nWe support:\nsotw non-aggregated\nsotw aggregated\nincremental non-aggregated\nincremental aggregated\n", variant)
		}
	}
	return nil, supported
}

func combineTags(godogTags string, customTags []string) (tags string) {
	if godogTags != "" {
		customTags = append(customTags, godogTags)
	}
	tags = strings.Join(customTags, " && ")
	return tags
}

func main() {
	pflag.Parse()
	godogOpts.Paths = pflag.Args()
	variantsInConfig := []string{}

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if *config != ""{
		c, err := yaml.ReadFile(*config)
		if err != nil {
			log.Info().Msgf("Error reading config file %v: %v\n", *config, err)
		}
		target, err := c.Get("targetAddress")
		if err != nil {
			log.Info().Msgf("Cannot get target address from config file: %v\n", err)
		} else {
		  *targetAddress = target
		}
		adapter, err := c.Get("adapterAddress")
		if err != nil {
			log.Info().Msgf("Cannot get adapter address from config file: %v\n", err)
		} else {
			*adapterAddress = adapter
		}
		variants, err := yaml.Child(c.Root, "variants")
		if err != nil {
			log.Info().Msgf("Error getting variants from config: %v\n", err)
		}
		varList, ok := variants.(yaml.List)
		if ok {
			for i := 0; i < varList.Len(); i++ {
				node := varList.Item(i)
				variant := string(node.(yaml.Scalar))
				variantsInConfig = append(variantsInConfig, variant)
			}
		}
	}

	suite := godog.TestSuite{
		Name:                 "xDS Test Suite",
		ScenarioInitializer:  InitializeScenario,
		TestSuiteInitializer: InitializeTestSuite,
		Options:              &godogOpts,
	}

	// any tags passed in with -t when invoking the runner
	godogTags := godogOpts.Tags
	supportedVariants := make(map[string]bool)
	if len(variantsInConfig) > 0 {
		_, supportedVariants = parseSupportedVariants(variantsInConfig)
	} else {
	   _, supportedVariants = parseSupportedVariants(*variant)
	}

	if supportedVariants["sotw non-aggregated"] {
		incremental = false
		aggregated = false
		customTags := []string{"@sotw", "@non-aggregated"}
		godogOpts.Tags = combineTags(godogTags, customTags)
		suite.Run()
	}

	if supportedVariants["sotw aggregated"] {
		incremental = false
		aggregated = true
		customTags := []string{"@sotw", "@aggregated"}
		godogOpts.Tags = combineTags(godogTags, customTags)
		suite.Run()
	}

	if supportedVariants["incremental non-aggregated"] {
		incremental = true
		aggregated = false
		customTags := []string{"@incremental", "@non-aggregated"}
		godogOpts.Tags = combineTags(godogTags, customTags)
		suite.Run()
	}

	if supportedVariants["incremental aggregated"] {
		incremental = true
		aggregated = true
		customTags := []string{"@incremental", "@aggregated"}
		godogOpts.Tags = combineTags(godogTags, customTags)
		suite.Run()
	}
	os.Exit(0)
}
