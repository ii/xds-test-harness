package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cucumber/godog"
	"github.com/ii/xds-test-harness/internal/parser"
	"github.com/ii/xds-test-harness/internal/runner"
	"github.com/ii/xds-test-harness/internal/types"
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
	godogOpts      = godog.Options{}
)

func init() {
	godog.BindCommandLineFlags("godog.", &godogOpts)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	pflag.Parse()
	godogTags := godogOpts.Tags

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// default to using CLI Flag for settings
	err, supportedVariants := parser.ParseSupportedVariants(*variant)
	if err != nil {
		log.Fatal().Msgf("Cannot parse variants from CLI: %v\n", err)
	}
	// If config present, use it for all non-debugging values
	if *config != "" {
		*targetAddress, *adapterAddress, *nodeID, supportedVariants = parser.ValuesFromConfig(*config)
	}

	var results types.Results
	for _, variant := range supportedVariants {
		log.Info().
			Msgf("Starting Tests for %v\n", string(variant))

		suite := runner.NewSuite(variant)
		err, variantResults := suite.Run(*adapterAddress, *targetAddress, *nodeID, godogTags)
		if err != nil {
			log.Fatal().
				Msgf("Error when attempting to run test suite: %v\n", err)
		}
		results = runner.UpdateResults(results, variantResults)
	}
	log.Info().
		Msgf("All done, here are your results!!: %v\n", results)

	file, _ := json.MarshalIndent(results, "", "  ")
	_ = ioutil.WriteFile("results.json", file, 0644)
	os.Exit(0)
}
