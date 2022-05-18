package main

import (
	"encoding/json"
	"fmt"
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
	testWriting    = pflag.BoolP("testwriting", "W", false, "Sets a pretty output that doesn't write to file, for better feedback while writing tests.")
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
			Msgf("Starting Tests for %v", string(variant))

		suite := runner.NewSuite(variant, *testWriting)
		if err = suite.StartRunner(*nodeID, *adapterAddress, *targetAddress); err != nil {
			log.Fatal().
				Err(err).
				Msg("Could not start runner.")
		}
		if err = suite.SetTags(godogTags); err != nil {
			log.Fatal().
				Err(err).
				Msg("Could not set tags properly to start up test suite.")
		}
		suite.ConfigureSuite()

		variantResults, err := suite.Run()
		if err != nil {
			log.Fatal().
				Msgf("Error when attempting to run test suite: %v\n", err)
		}
		results = runner.UpdateResults(results, variantResults)
	}
	if !*testWriting {
		printResults(results)
		file, _ := json.MarshalIndent(results, "", "  ")
		_ = ioutil.WriteFile("results.json", file, 0644)
	}
	os.Exit(0)
}

func printResults(results types.Results) {

	divider := "-------------------"
	fmt.Println("\nTest Suite Finished\n" + divider)
	fmt.Printf("Ran %v tests across %v variants\n\n", results.Total, len(results.Variants))
	fmt.Println("Passed: ", results.Passed)
	if results.Failed > 0 {
		fmt.Println("Failed: ", results.Failed)
	}
	if results.Skipped > 0 {
		fmt.Println("Skipped: ", results.Skipped)
	}
	if results.Undefined > 0 {
		fmt.Println("Undefined: ", results.Undefined)
	}
	if results.Pending > 0 {
		fmt.Println("Pending: ", results.Pending)
	}
	fmt.Printf("\n\nResults broken down by Variant....\n\n")
	for _, variant := range results.ResultsByVariant {
		if variant.Total > 0 {
			fmt.Println(variant.Name + "\n" + divider)
			fmt.Println(variantResults(variant))
		}
	}
}

func variantResults(results types.VariantResults) string {
	total := fmt.Sprintf("Total tests: %v\n", results.Total)
	passed := fmt.Sprintf("Passed: %v\n", results.Passed)
	failed := fmt.Sprintf("Failed: %v\n", results.Failed)
	skipped := fmt.Sprintf("Skipped: %v\n", results.Skipped)
	undefined := fmt.Sprintf("Undefined: %v\n", results.Undefined)
	pending := fmt.Sprintf("Pending: %v\n", results.Pending)
	var failedTests string
	if len(results.FailedScenarios) > 0 {
		failedTests = "Failed Tests:\n"
		for _, test := range results.FailedScenarios {
			failedTests = failedTests + "  - " + test.Name + "\n"
		}
	}
	return total + passed + failed + failedTests + skipped + undefined + pending
}
