package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/ii/xds-test-harness/internal/types"
	"github.com/rs/zerolog/log"
)

func init() {
	godog.Format("xds", "Progress formatter with emojis", xdsFormatterFunc)
}

func xdsFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newxdsFmt(suite, out)
}

type xdsFmt struct {
	*godog.ProgressFmt
	out     io.Writer
	results types.VariantResults
}

func newxdsFmt(suite string, out io.Writer) *xdsFmt {
	return &xdsFmt{
		ProgressFmt: godog.NewProgressFmt(suite, out),
		out:         out,
		results:     types.VariantResults{},
	}
}

func (f *xdsFmt) Passed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Passed(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.results.Total++
	f.results.Passed++
	f.step(step.Id, scenario)
}

func (f *xdsFmt) Skipped(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Skipped(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.results.Total++
	f.results.Skipped++
	f.step(step.Id, scenario)
}

func (f *xdsFmt) Undefined(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Undefined(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.results.Total++
	f.results.Undefined++
	f.step(step.Id, scenario)
}

func (f *xdsFmt) Failed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition, err error) {
	f.ProgressFmt.Base.Failed(scenario, step, match, err)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.results.Total++
	f.results.Failed++
	f.step(step.Id, scenario)
}

func (f *xdsFmt) Pending(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Pending(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.results.Total++
	f.results.Pending++
	f.step(step.Id, scenario)
}

func (f *xdsFmt) Summary() {
	f.results.FailedScenarios = f.gatherFailedScenarios()
	data, err := json.MarshalIndent(f.results, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(f.out, "%s\n", string(data))
}

func (f *xdsFmt) step(pickleStepID string, scenario *godog.Scenario) {
	pickleStepResult := f.Storage.MustGetPickleStepResult(pickleStepID)
	printStatusEmoji(pickleStepResult.Status)

	lastStep := isLastStep(pickleStepID, scenario)
	if lastStep {
		failed, failedStep := f.didScenarioFail(scenario)
		if failed {
			log.Info().
				Str("failed step", failedStep).
				Msgf("| [%v]%v", colors.Red("FAILED"), scenario.Name)
		} else {
			log.Info().
				Msgf("| [%v]%v", colors.Green("PASSED"), scenario.Name)
		}
	}
	*f.Steps++
}

func (f *xdsFmt) didScenarioFail(scenario *godog.Scenario) (failed bool, failedStep string) {
	failed = false
	results := f.Storage.MustGetPickleStepResultsByPickleID(scenario.Id)
	for _, result := range results {
		if result.Status.String() == "failed" {
			feature := f.Storage.MustGetFeature(scenario.Uri)
			pickleStep := f.Storage.MustGetPickleStep(result.PickleStepID)
			step := feature.FindStep(pickleStep.AstNodeIds[0])
			failed = true
			failedStep = step.Text
		}
	}
	return failed, failedStep
}

func (f *xdsFmt) gatherFailedScenarios() (failedScenarios []types.FailedScenario) {
	failedSteps := f.Storage.MustGetPickleStepResultsByStatus(1)
	for _, failure := range failedSteps {
		scenario := f.Storage.MustGetPickle(failure.PickleID)
		feature := f.Storage.MustGetFeature(scenario.Uri)
		pickleStep := f.Storage.MustGetPickleStep(failure.PickleStepID)
		step := feature.FindStep(pickleStep.AstNodeIds[0])

		fs := types.FailedScenario{
			Name:       scenario.Name,
			FailedStep: step.Text,
			Line:       fmt.Sprintf("%v:%v", feature.Uri, step.Location.Line),
		}

		failedScenarios = append(failedScenarios, fs)
	}
	return failedScenarios
}

func printStatusEmoji(status godog.StepResultStatus) {
	switch status {
	case godog.StepPassed:
		fmt.Printf(" %s", "✅")
	case godog.StepSkipped:
		fmt.Printf(" %s", "➖")
	case godog.StepFailed:
		fmt.Printf(" %s", "❌")
	case godog.StepUndefined:
		fmt.Printf(" %s", "❓")
	case godog.StepPending:
		fmt.Printf(" %s", "🚧")
	}
}

func isLastStep(stepId string, scenario *godog.Scenario) bool {
	stepIds := []string{}
	lastIndex := len(scenario.Steps) - 1
	for _, step := range scenario.Steps {
		stepIds = append(stepIds, step.Id)
	}
	index := indexOf(stepId, stepIds)
	return index == lastIndex
}

func indexOf(val string, arr []string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}
