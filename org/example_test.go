package main

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress", // can define default values
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	log.Println("TestMain is being called with Testing.M")
	// log.Println("TestMain is being called with Testing.M: %v", m.tests)
	flag.Parse()
	opts.Paths = flag.Args()

	status := godog.TestSuite{
		Name:                 "godogs",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}.Run()

	// Optional: Run `testing` package's logic besides godog.
	if st := m.Run(); st > status {
		status = st
	}

	os.Exit(status)

}

// type testData struct {
// 	status
// }

// func TestingMainFeatureContext(s *godog.Suite, m *testing.M) {
// 	e2elog.Logf("Adding Before Suite")
// 	FirstStepsFeatureContext(s, m)
// }

// func FirstStepsFeatureContext(s *godog.Suite, m *testing.M) {
// 	s.Step(`^a change is made$`, aChangeIsMade)
// 	s.Step(`^a starting point$`, aStartingPoint)
// 	s.Step(`^we see results$`, weSeeResults)
// 	// testingMain = m
// 	// s.BeforeScenario(func(interface{})) {}
// }

func aChangeIsMade() error {
	log.Println("A Change was Made!!")
	return nil
}

func aStartingPoint() error {
	log.Println("We configured a starting point!!")
	return nil
}

func weSeeResults() error {
	log.Println("We canhaz results!!")
	return nil
}
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	log.Println("Adding BeforeSuite Func()")
	ctx.BeforeSuite(func() {
		log.Println("Running BeforeSuite()")
	})
	log.Println("Adding AfterSuite Func()")
	ctx.AfterSuite(func() {
		log.Println("Running AfterSuite()")
	})
}
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(sc *godog.Scenario) {
		log.Printf("We are BeforeScenario : %#v : %#v : %#v\n",
			sc.Uri,  // "example.feature"
			sc.Name, // ""
			sc.Id,   // "7"
		)
		log.Printf("Settings Godogs = 0")
		// ./example_test.go:93:3: undefined: Godogs
		// How do we set godocs?
	})
	ctx.AfterScenario(func(sc *godog.Scenario, err error) {
		log.Printf("We are AfterScenario : %#v : %#v : %#v\n",
			sc.Uri,  // "example.feature"
			sc.Name, // ""
			sc.Id,   // "7"
		)
		if err != nil {
			log.Printf("AfterScenario err : %#v", err)
			// log.Printf("AfterScenario err : %#v", err.errorString)
		}
	})
	ctx.BeforeStep(func(st *godog.Step) {
		log.Printf("We are BeforeStep : %#v : %#v\n", st.Id, st.Text)
	})
	ctx.AfterStep(func(st *godog.Step, err error) {
		log.Printf("We are AfterStep : %#v : %#v\n", st.Id, st.Text)
		if err != nil {
			log.Printf("AfterStep err : %#v", err)
			// log.Printf("AfterStep err : %#v", err.errorString)
		}
	})
	ctx.Step(`^a starting point$`, aStartingPoint)
	ctx.Step(`^a change is made$`, aChangeIsMade)
	ctx.Step(`^we see results$`, weSeeResults)
}

// assertExpectedAndActual is a helper function to allow the step function to call
// assertion functions where you want to compare an expected and an actual value.
func assertExpectedAndActual(a expectedAndActualAssertion, expected, actual interface{}, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, expected, actual, msgAndArgs...)
	return t.err
}

type expectedAndActualAssertion func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool

// assertActual is a helper function to allow the step function to call
// assertion functions where you want to compare an actual value to a
// predined state like nil, empty or true/false.
func assertActual(a actualAssertion, actual interface{}, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, actual, msgAndArgs...)
	return t.err
}

type actualAssertion func(t assert.TestingT, actual interface{}, msgAndArgs ...interface{}) bool

// asserter is used to be able to retrieve the error reported by the called assertion
type asserter struct {
	err error
}

// Errorf is used by the called assertion to report an error
func (a *asserter) Errorf(format string, args ...interface{}) {
	a.err = fmt.Errorf(format, args...)
}
