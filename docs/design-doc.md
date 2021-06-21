- [Context and Scope](#orga28b9eb)
  - [Goals](#orgfb2ae8f)
  - [Non-goals](#org12df928)
- [Overview](#org4d4bad9)
- [Concepts and Definitions](#org71bc43e)
  - [Behaviour Driven Design](#orga0f94d7)
  - [Cucumber](#orgb25d1f1)
  - [Gherkin](#org21684cf)
  - [Features](#org0a6f7ea)
  - [Scenarios](#org7c596de)
  - [Steps](#orgced4020)
  - [Godog](#orgae74c25)
- [Detailed Design](#org84d6190)
  - [Architecture](#org872d6ce)
  - [Test Case Format](#org9ad761d)
    - [Example](#orga60bb47)
  - [The Runner](#orgaebfd14)
  - [The Adapter protocol](#orge5b8772)
    - [Scope](#org7e45653)
    - [Spec](#org1d6c97b)
- [Considered Alternatives](#org9c789e2)



<a id="orga28b9eb"></a>

# Context and Scope

This document outlines the implementation of a conformance test suite for the xDS protocol. It builds off designs and requirements established in the [xDS conformance Statement of Work](https://docs.google.com/document/d/17E3k4fGJedVISCudrW4Kgzf89gvIIhAdZnJmo6pMVlA/edit), outlining how these requirements can be best implemented.


<a id="orgfb2ae8f"></a>

## Goals

In addition to the goals outlined in the SoW, our implementation is intended to achieve the following:

-   An extensible test case syntax behaviour that tests *behaviour*.
-   The ability to write new tests without writing new code.
-   A framework that is easy to contribute to.
-   A framework that is flexible enough to handle new requirements as discovered.


<a id="org12df928"></a>

## Non-goals

-   Write a complete test case language.
-   Include tests that cover envoy specific scenarios (the cases should be implementation-agnostic).
-   Write all known test cases.


<a id="org4d4bad9"></a>

# Overview

The test suite is a collection of behaviour-driven tests, a test runner that implements these tests in code and runs them against a target, and an adapter spec implemented in the target so our runner can set up appropriate state.

In use, the runner is a language-agnostic binary that a team would run against their xDS server to validate whether their system is conformant. It is expected that the team had implemented the adapter in their language and style prior to running the tests. After running the binary, the team would have an xml file with their complete results.

The test cases are written separate from their implementation, using the [gherkin](https://cucumber.io/docs/gherkin/reference/) testing language. This allows for a consistency in how test scenarios are articulated, and the ability to reuse test steps, minimizing the amount of new code needing to be written.


<a id="org71bc43e"></a>

# Concepts and Definitions


<a id="orga0f94d7"></a>

## Behaviour Driven Design

The tests are written to support &ldquo;Behaviour Driven Design&rdquo;, or BDD. BDD is a communication methodology that seeks to create a common language between stakeholders and developers/test-writers. A deeper introduction can be read here: [Intro to BDD](https://cucumber.io/docs/bdd/).

In practice, it means our tests are first written in a human-readable(near natural language) syntax which describes, at a high level, the conformant features of xDS. The feature is then mapped to a test implementation that then checks whether the target adequately fulfills the expected behaviour.

A core part of this approach is that a behaviour is written as a series of steps,t he steps following an agreed-upon syntax. The steps are then mapped to their corresponding function. This means that new tests can be written using the same syntax as prior ones, re-using the corresponding functions. This allows for people to write new tests without having to write new code, or without having knowledge of the underlying code itself.


<a id="orgb25d1f1"></a>

## Cucumber

[Cucumber](https://cucumber.io/docs/guides/overview/) is a formalization of BDD concepts into a language and testing grammar. It uses a syntax named `gherkin` to specify expected behaviour in such a way that can be easily translated into tests.


<a id="org21684cf"></a>

## Gherkin

[Gherkin](https://cucumber.io/docs/gherkin/) is the testing/spec language of Cucumber. It looks like this:

    Feature: Buying Vegetables
      Scenario: Buying vegetables reduces their quantity
        Given there are 12 courgettes.
        When a shopper purchases 5 of them.
        Then there are 7 courgettes remaining.

Notably, there is a small set of *keywords* that each line starts with that are used to map each step to a testable function.

Gherkin tests are organized around a model of scenarios, steps, and features.


<a id="org0a6f7ea"></a>

## Features

A behaviour of the system described at the highest level. Most often, the feature has multiple scenarios that, in toto, give a full description of this intended behaviour.

For example: &ldquo;Subscriptions&rdquo; might be a feature of the xDS protocol, that can be described through example scenarios like: subscribe, unsusbcribe, and &ldquo;unsubscribe then resubscribe&rdquo;.

Practically this means our test suite is set up as a collection of feature files, with each feature being a collection of scenarios.


<a id="org7c596de"></a>

## Scenarios

Clear, simple, self-contained depictions of some aspect of an intended feature. These are the heart of Cucumber and our testing suite, and are the closest analogy to a test. A scenario is structured as:

-   Given some state
-   When an action occurs
-   Then resulting state can be observed.

Scenarios ar emeant to be declarative and not a line by line specification of how a test operates. Each line of a scenario is called a **step**.


<a id="orgced4020"></a>

## Steps

A line in a scenario, written in gherkin syntax as a Keyword followed by some natural language description. Steps can also include dynamic variables and code blocks. If a step is written cleanly enough, it can be used in multiple scenarios. For example, take a step written like so:

    Given a server with state matching yaml:
    ```
    ...some specific yaml...
    ```

This would be mapped to a test function where the yaml is a parameter passed in. This means you could have another scenario testing some different state, with different yaml, but using the same function.


<a id="orgae74c25"></a>

## Godog

[Godog](https://github.com/cucumber/godog#godog) is a library for setting up a test suite from gherkin feature files. It is the core of our test suite, used to build up the framework and iterate through all our tests.


<a id="org84d6190"></a>

# Detailed Design


<a id="org872d6ce"></a>

## Architecture

The architecture closely matches the original diagram in the Statement of Work.

Our test binary is invoked with a simple configuration specifying the address of the target and its adapter. The suite starts up an instance of our test runner then iterates over a collection of feature files, running the matching test function for each scenario step.

The test functions utilize the target&rsquo;s adapter implementation to setup any necessary state on the target, communicates to the taget directly via xDS, and passes along streams and state from step to step via the Runner.

The results of each test are output to a local junit.xml for further sturdy or, potentially, certification.

![img](./assets/architecture.png)

The crucial aspects for us to implement in this design is a clean and consistent syntax for our tests and a strong, flexible runner.


<a id="org9ad761d"></a>

## Test Case Format

The test repo will be organized, roughly, like so:

    ../../test-suite
    ├── features
    │   ├── subscriptions.feature
    │   └── warming.feature
    ├── internal
    │   ├── parser
    │   └── runner
    ├── main.go
    └── steps
        ├── common.go
        ├── subscriptions.go
        └── warming.go
    
    5 directories, 6 files

Importantly, for the test writer, the tests are specified as a `.feature` file, and then implemented using a combination of common and feature-specific steps that are implemented on the runner.

One goal of the project is, when there is a new feature, 60% of its scenarios steps can be described using existing steps. This reduces the volume of new code needing to be written, and allows for contributions from people who know xDS well, but do not need to know golang or the details of our implementation.


<a id="orga60bb47"></a>

### Example

Let&rsquo;s say in the future, some new feature of xDS is introduced that needs to be tested. For simplicity sake, let&rsquo;s give the feature some random name like &ldquo;jumproping&rdquo;.

A test writer wants to implement new tests in a PR to be merged into our test suite. In this example, the test writer is proficient in golang.

To start, they&rsquo;d write up a features file at `features/jumprope.feature` This would describe the jumproping feature at 10,000ft, illustrated with a set of scenarios for each aspect of it.

The scnearios would contain steps pulled from the common library: setting state, passing along messages, validating responses, etc. In addition, there are some aspects of jumproping not covered in our common library.

They implement these new steps in golang in `steps/jumprope.go`

Lastly, they add these new steps toa collection of them in our `main.go`. The order of placement is not important, as the steps are mapped to the scenarios via regex.

Their pr would include changes to these three files: `features/jumprope.go`, `steps/jumprope.go`, and `main.go`.

Later on, nuances are found within jumproping that need to have their own tests. A test-writer, without golang proficiency, reads through the `jumprope.feature` and the documentation of common steps, and writes a new set of scenarios built from existing steps. They open a new PR with changes only to `features/jumprope.go`.


<a id="orgaebfd14"></a>

## The Runner

All of the steps in our `steps` folder are test functions implemented as methods to our Runner struct. This Runner holds state that is meant to be passed from step to step.

For example, a basic example of a runner might be:

    type ClientConfig struct {
    	Port string
    	Conn *grpc.ClientConn
    }
    
    type Runner struct {
    	Adapter           *ClientConfig
    	Target            *ClientConfig
    	CDS               struct {
    		Stream    cluster_service.ClusterDiscoveryService_StreamClustersClient
    		Responses []*envoy_service_discovery_v3.DiscoveryResponse
    	}
    }

Before a scenario, the runner would connect to the target and adapter via gRPC. It would store these connections to be accessed by each of the steps within the scenario. Similarly, one step may invoke a discovery response that is stored in the runner then validated in the following step. A hook is run after a scenario that maintains the adapter and target connections, but cleans out any other state.


<a id="orge5b8772"></a>

## The Adapter protocol

The adapter is a gRPC API defined in the test harness repo whose intention is to set the required state for each test to run cleanly in isolation.


<a id="org7e45653"></a>

### Scope

The adapter is meant to be simple and limited. It can `SetState`, fully resetting the target to some specified beginning state. Or, it can `UpdateState` when the scenario is stateful, e.g. we want to track the chain of versions created across each step.


<a id="org1d6c97b"></a>

### Spec

The adapter api is specified within the test-suite repo as a protobuf schema. It is the responsibility of the test target scheme to implement the adapter in the language and style of their server.

A basic version of the schema can be found in this repo&rsquo;s [api/adapter.proto](https://github.com/ii/xds-test-harness/blob/design-doc/api/adapter/adapter.proto#L65)


<a id="org9c789e2"></a>

# Considered Alternatives

The tests could be written in a different style or syntax, for example using the native go testing funtionality of golang and something like [ginkgo](https://onsi.github.io/ginkgo/). However, it is important to us that the tests can be read, discussed, and reasoned about by anyone with xDS knowledge, without having to know our implementation. Any testing syntax that was intertangled with a programming language was a non-starter for us.

Cucumber/gherkin provided the cleanest, and most established method for these BDD tests.

Cucumber provides a grammar and syntax, but doesn&rsquo;t specify how these feature files should be converted into tests. For this, we could either build an in-house solution or use an existing library. For simplicity, and stability, we chose the existing godog library. It is written by the cucumber team, is well documented, and provided a lot of happiness while working with it in our initial proof-of-concept.
