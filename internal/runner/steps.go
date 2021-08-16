package runner

import (
	"github.com/cucumber/godog"
)

func (r *Runner) LoadSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^a target setup with the following clusters:$`, r.ATargetSetupWithTheFollowingClusters)
	ctx.Step(`^a target setup with the following "([^"]*)":$`, r.ATargetSetupWithTheFollowing)
	ctx.Step(`^the Runner receives the following "([^"]*)":$`, r.TheRunnerReceivesTheFollowing)
	ctx.Step(`^the Runner sends a CDS wildcard request$`, r.TheRunnerSendsACDSWildcardRequest)
	ctx.Step(`^the Runner says "([^"]*)"$`, r.TheRunnerSays)
}

func (r *Runner) ATargetSetupWithTheFollowingClusters(arg1 *godog.DocString) error {
	return nil
}

func (r *Runner) ATargetSetupWithTheFollowing(arg1 string, arg2 *godog.DocString) error {
	return godog.ErrPending
}

func (r *Runner) TheRunnerReceivesTheFollowing(arg1 string, arg2 *godog.DocString) error {
	return godog.ErrPending
}

func (r *Runner) TheRunnerSendsACDSWildcardRequest() error {
	return godog.ErrPending
}

func (r *Runner) TheRunnerSays(arg1 string) error {
	return godog.ErrPending
}
