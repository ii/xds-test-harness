// Test suite builders for each of the variants.
package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/ii/xds-test-harness/internal/types"
	"github.com/rs/zerolog/log"
)

type Suite struct {
	Variant     types.Variant
	Runner      *Runner
	Aggregated  bool
	Incremental bool
	Buffer      bytes.Buffer
	Tags        string
	TestSuite   godog.TestSuite
}

func (s *Suite) StartRunner(node, adapter, target string) error {
	s.Runner = FreshRunner()
	s.Runner.NodeID = node
	s.Runner.Aggregated = s.Aggregated

	if err := s.Runner.ConnectClient("target", target); err != nil {
		log.Fatal().
			Msgf("error connecting to target: %v", err)
	}
	if err := s.Runner.ConnectClient("adapter", adapter); err != nil {
		log.Fatal().
			Msgf("error connecting to adapter: %v", err)
	}

	log.Info().
		Msgf("Connected to target at %s and adapter at %s\n", target, adapter)

	if s.Runner.Aggregated {
		log.Info().
			Msgf("Tests will be run via a single aggregated streams")
	} else {
		log.Info().
			Msgf("Tests will be run via separate, non-aggregated streams")
	}
	return nil
}

func (s *Suite) SetTags(base string) error {
	tagList := []string{}
	variants := strings.Split(string(s.Variant), " ")

	if len(variants) < 1 {
		err := fmt.Errorf("No variant type found to create tags from. This means the suite was not initialized properly.")
		return err
	}
	for _, tag := range variants {
		tag = "@" + tag
		tagList = append(tagList, tag)
	}
	if base != "" {
		tagList = append(tagList, base)
	}
	s.Tags = strings.Join(tagList, " && ")
	return nil
}

func (b *Suite) ConfigureSuite() error {
	initScenario := func(ctx *godog.ScenarioContext) {
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			log.Debug().Msg("Creating Fresh Runner!")
			b.Runner = FreshRunner(b.Runner)
			return ctx, nil
		})
		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			c := pb.NewAdapterClient(b.Runner.Adapter.Conn)
			clearRequest := &pb.ClearRequest{Node: b.Runner.NodeID}
			clear, err := c.ClearState(context.Background(), clearRequest)
			if err != nil {
				log.Err(err).
					Msg("Couldn't clear state")
			}
			log.Debug().
				Msgf("Clearing State: %v\n", clear.Response)
			return ctx, nil
		})
		b.Runner.LoadSteps(ctx)
	}

	godogOpts := godog.Options{
		ShowStepDefinitions: false,
		Randomize:           0,
		StopOnFailure:       false,
		Strict:              false,
		NoColors:            false,
		Tags:                b.Tags,
		Format:              "xds",
		Output:              &b.Buffer,
		Concurrency:         0,
	}

	suite := godog.TestSuite{
		Name:                fmt.Sprintf("xds Test Suite [%v]", b.Variant),
		ScenarioInitializer: initScenario,
		Options:             &godogOpts,
	}

	b.TestSuite = suite
	return nil
}

func (s *Suite) Run(adapter,target,nodeId,tags string) (err error, results types.VariantResults) {
	if err = s.StartRunner(nodeId,adapter,target); err != nil {
		return
	}
	if err = s.SetTags(tags); err != nil {
		return
	}
	if err = s.ConfigureSuite(); err != nil {
		return
	}

	s.TestSuite.Run()

	cukeResults := []types.CukeFeatureJSON{}
	if err = json.Unmarshal([]byte(s.Buffer.String()), &cukeResults); err != nil {
		err = fmt.Errorf("Error unmarshalling test results: %v\n", err)
		return err, results
	}

	for _, cuke := range cukeResults {
		results = gatherResults(results, cuke)
	}

	results.Name = string(s.Variant)
	return err, results
}

func NewSotwNonAggregatedSuite() *Suite {
	return &Suite{
		Variant:     types.SotwNonAggregated,
		Aggregated:  false,
		Incremental: false,
		Buffer:      *bytes.NewBuffer(nil),
	}
}

func NewSotwAggregatedSuite() *Suite {
	return &Suite{
		Variant:     types.SotwAggregated,
		Aggregated:  true,
		Incremental: false,
		Buffer:      *bytes.NewBuffer(nil),
	}

}

func NewIncrementalNonAggregatedSuite() *Suite {
	return &Suite{
		Variant:     types.IncrementalNonAggregated,
		Aggregated:  false,
		Incremental: true,
		Buffer:      *bytes.NewBuffer(nil),
	}

}

func NewIncrementalAggregatedSuite() *Suite {
	return &Suite{
		Variant:     types.IncrementalAggregated,
		Aggregated:  true,
		Incremental: true,
		Buffer:      *bytes.NewBuffer(nil),
	}
}

func NewSuite(variant types.Variant) *Suite {
	switch variant {
	case types.SotwNonAggregated:
		return NewSotwNonAggregatedSuite()
	case types.SotwAggregated:
		return NewSotwAggregatedSuite()
	case types.IncrementalNonAggregated:
		return NewIncrementalNonAggregatedSuite()
	case types.IncrementalAggregated:
		return NewIncrementalAggregatedSuite()
	default:
		return nil
	}
}


func UpdateResults(current types.Results, variantResults types.VariantResults) types.Results {
	return types.Results{
		Total:            current.Total  + variantResults.Total,
		Passed:           current.Passed + variantResults.Passed,
		Failed:           current.Failed + variantResults.Failed,
		Variants:         append(current.Variants, variantResults.Name),
		ResultsByVariant: append(current.ResultsByVariant, variantResults),
	}
}

func gatherResults(current types.VariantResults, cuke types.CukeFeatureJSON) types.VariantResults {
	totalTests := len(cuke.Elements)
	passed := 0
	failed := 0
	failedTests := []types.FailedTest{}

	for _, test := range cuke.Elements {
		testPassed := true
		for _, step := range test.Steps {
			if step.Result.Status == "failed" {
				testPassed = false
				failedTests = append(failedTests, createFailedTest(test, step))
			}
		}
		if testPassed {
			passed++
		} else {
			failed++
		}
	}
	current.Total += int64(totalTests)
	current.Passed += int64(passed)
	current.Failed += int64(failed)
	current.FailedTests = append(current.FailedTests, failedTests...)
	return current
}

func createFailedTest(scenario types.CukeElement, failedStep types.CukeStep) types.FailedTest {
	return types.FailedTest{
		Scenario:   scenario.Name,
		FailedStep: failedStep.Name,
		Source:     failedStep.Match.Location,
	}
}

type SuiteConfig struct {
	adapter string
	target  string
	nodeId  string
	tags    string
}
