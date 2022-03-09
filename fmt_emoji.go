package main

import (
	"fmt"
	"io"
	"encoding/json"
	"math"

	"github.com/ii/xds-test-harness/internal/types"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/rs/zerolog/log"
)

const (
	passedEmoji    = "✅"
	skippedEmoji   = "➖"
	failedEmoji    = "❌"
	undefinedEmoji = "❓"
	pendingEmoji   = "🚧"
)

var (
	red   = colors.Red
	green = colors.Green
)

func init() {
	godog.Format("emoji", "Progress formatter with emojis", emojiFormatterFunc)
}

func emojiFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newEmojiFmt(suite, out)
}

type emojiFmt struct {
	*godog.ProgressFmt

	out io.Writer
}

func newEmojiFmt(suite string, out io.Writer) *emojiFmt {
	return &emojiFmt{
		ProgressFmt: godog.NewProgressFmt(suite, out),
		out:         out,
	}
}

type StepOrder int
const (
	First StepOrder = iota
	Middle
	Last
)

type StepResultStatus int
const (
	// order based on godog's internal model.StepResultStatus
	// but it is internal, and so cannot be used.
	Passed StepResultStatus = iota
	Failed
	Skipped
	Undefined
	Pending
)

// func (f *emojiFmt) TestRunStarted() {}

func (f *emojiFmt) Passed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Passed(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.step(step.Id, scenario)
}

func (f *emojiFmt) Skipped(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Skipped(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.step(step.Id, scenario)
}

func (f *emojiFmt) Undefined(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Undefined(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.step(step.Id, scenario)
}

func (f *emojiFmt) Failed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition, err error) {
	f.ProgressFmt.Base.Failed(scenario, step, match, err)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.step(step.Id, scenario)
}

func (f *emojiFmt) Pending(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Pending(scenario, step, match)
	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()
	f.step(step.Id, scenario)
}

func (f *emojiFmt) Summary() {
	results := &types.VariantResults{}
	results.Passed = f.countByStatus(Passed)
	results.Failed = f.countByStatus(Failed)
	results.Skipped = f.countByStatus(Skipped)
	results.Undefined = f.countByStatus(Undefined)
	results.Pending = f.countByStatus(Pending)
	results.FailedScenarios = f.gatherFailedScenarios()
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	fmt.Fprintf(f.out, "%s\n",string(data))
}

func (f *emojiFmt) step(pickleStepID string, scenario *godog.Scenario) {
	pickleStepResult := f.Storage.MustGetPickleStepResult(pickleStepID)
	position := scenarioPosition(pickleStepID, scenario)
	if position == Last {
		passed := true
		var failedStep string
		results := f.Storage.MustGetPickleStepResultsByPickleID(scenario.Id)
		for _, result := range results {
			r := fmt.Sprintf("%v", result.Status)
			if r == "failed" {
				passed = false
				failedStep = "this is the failed step"
				break
			}
		}
		if passed {
			log.Info().Msgf("[%v]%v", green("PASSED"), scenario.Name)
		} else {
			log.Info().Str("failed step", failedStep).Msgf("[%v]%v", red("FAILED"), scenario.Name)
		}
	} else {
		switch pickleStepResult.Status {
		case godog.StepPassed:
			fmt.Printf(" %s", passedEmoji)
		case godog.StepSkipped:
			fmt.Printf(" %s", skippedEmoji)
		case godog.StepFailed:
			fmt.Printf(" %s", failedEmoji)
		case godog.StepUndefined:
			fmt.Printf(" %s", undefinedEmoji)
		case godog.StepPending:
			fmt.Printf(" %s", pendingEmoji)
		}

	}
	*f.Steps++

	if math.Mod(float64(*f.Steps), float64(f.StepsPerRow)) == 0 {
		fmt.Fprintf(f.out, " %d\n", *f.Steps)
	}
}

func (f *emojiFmt) countByStatus(status StepResultStatus) int {
	switch status {
	case Passed:
		return len(f.Storage.MustGetPickleStepResultsByStatus(0))
	case Failed:
		return len(f.Storage.MustGetPickleStepResultsByStatus(1))
	case Skipped:
		return len(f.Storage.MustGetPickleStepResultsByStatus(2))
	case Undefined:
		return len(f.Storage.MustGetPickleStepResultsByStatus(3))
	case Pending:
		return len(f.Storage.MustGetPickleStepResultsByStatus(4))
	default:
		return 0
	}
}

func (f *emojiFmt) gatherFailedScenarios () (failedScenarios []types.FailedScenario) {
	failedSteps := f.Storage.MustGetPickleStepResultsByStatus(1)
	for _, failure := range failedSteps {
		scenario := f.Storage.MustGetPickle(failure.PickleID)
		feature := f.Storage.MustGetFeature(scenario.Uri)
		pickleStep := f.Storage.MustGetPickleStep(failure.PickleStepID)
		step := feature.FindStep(pickleStep.AstNodeIds[0])

		fs := types.FailedScenario{
			Name:       scenario.Name,
			FailedStep: step.Text,
			Line:       fmt.Sprintf("%v:%v",feature.Uri,step.Location.Line),
		}

		failedScenarios = append(failedScenarios, fs)
	}
	return failedScenarios
}

func scenarioPosition(stepId string, scenario *godog.Scenario) StepOrder {
	stepIds := []string{}
	lastIndex := len(scenario.Steps) - 1
	for _, step := range scenario.Steps {
		stepIds = append(stepIds, step.Id)
	}
	index := indexOf(stepId, stepIds)
	if index == 0 {
		return First
	}
	if index == lastIndex {
		return Last
	}
	return Middle

}

func indexOf(val string, arr []string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}
