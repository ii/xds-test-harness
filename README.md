# xDS Test Suite Prototype

A tool to test whether your xDS server is conformant to the xDS protocol.

*This repo is a _strong_ work in progress, and holds both working code and
learning experiments. We try to keep the two as separate and clear as possible.*

# Run the example

To see this in action, clone and navigate to this repo, and start a target server:

``` sh
git clone https://github.com/ii/xds-test-harness
cd xds-test-harness
go run example/go-control-plane/main/main.go
```

This starts up a simple implementation of the [envoy proxy go control
plane](https://github.com/envoyproxy/go-control-plane/) with an adapter
integration.

In another terminal, navigate to this repo and run the suite:
``` sh
cd xds-test-harness
go run .
```

You can also see detailed logging of the tests in action with

``` sh
go run . --debug
```

# Design

The suite is made up of tests, a main program to run these tests against a
target server, and an adapter for setting the server to the correct state for
each test. The adapter is meant to be integrated into the target, and so in this
repo is just the API. The full design can be read in our [design
doc](https://github.com/ii/xds-test-harness/blob/main/docs/design-doc.md)

# Navigating the repo
- All tests are held in
  [/features](https://github.com/ii/xds-test-harness/tree/main/features).
- `main.go` is the entrance to our program, but mostly calls the runner, located
  in [/internal](https://github.com/ii/xds-test-harness/tree/main/internal)
- the runner is made of two files,
  [runner.go](https://github.com/ii/xds-test-harness/blob/main/internal/runner/runner.go),
  which sets up the core mechanics and
  [steps.go](https://github.com/ii/xds-test-harness/blob/main/internal/runner/steps.go)
  which implements all the gherkin steps into go code.
- the adapter is outlined in
  [/api/adapter](https://github.com/ii/xds-test-harness/blob/main/api/adapter/adapter.proto).
  It is written as [protocol
  buffers](https://developers.google.com/protocol-buffers/docs/gotutorial)
- further documentation is held in
  [/docs](https://github.com/ii/xds-test-harness/tree/main/docs)
- notes, experiments, and other items of historical interest are held in
  [/diary](https://github.com/ii/xds-test-harness/tree/main/diary)

# Background and Context

This work is based off [xds Conformance Suite Statement of
Work](https://docs.google.com/document/d/17E3k4fGJedVISCudrW4Kgzf89gvIIhAdZnJmo6pMVlA/edit#heading=h.tqf1i1hfnem9)

To learn more about the xDS protocol, you can read [it's documentation on
envoyproxy.io](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)
or through the essay ["The Universal Data Plane
API"](https://blog.envoyproxy.io/the-universal-data-plane-api-d15cec7a)

The core testing functionality is handled by the excellent [godog
library](https://github.com/cucumber/godog)

The suite is built with and is testing gRPC services.
[grpc.io](https://grpc.io/) has a great [introduction to
grpc](https://grpc.io/docs/what-is-grpc/introduction/) and [tutorial for
go](https://grpc.io/docs/languages/go/basics/)
