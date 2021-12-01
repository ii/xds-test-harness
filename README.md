# xDS Test Suite Prototype

A tool to test whether your xDS server is conformant to the xDS protocol.

*This repo is a _strong_ work in progress, and holds both working code and
learning experiments. We try to keep the two as separate and clear as possible.*

# Run the example

You can run the test suite against our example server: an implementation of the go-control-plane that is integrated with our adapter.

To do this, you will **1** generate the api then **2** start the server and then **3** start the test suite

## Generate the API

The api uses [protocol
buffers](https://developers.google.com/protocol-buffers/), and so first you need
to install protoc:

[install instructions for protoc](https://grpc.io/docs/protoc-installation/).

You should be able to run this command and get a similar response:

``` sh
protoc --version
#=> returns libprotc 3.17.3+
```

Once installed, clone and navigate to this repo:

``` sh
git clone https://github.com/ii/xds-test-harness
cd xds-test-harness
```

and generate the api:

``` sh
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/adapter/adapter.proto
```

## Start the target server

In your terminal window, from this repo, run:

``` sh
go run examples/go-control-plane/main/main.go
```

## Run the suite

In a new terminal window, navigate to the harness and start it up:
``` sh
cd xds-test-harness
go run .
```

To run it with detailed logging, add the --debug flag:
``` sh
go run . --debug
```


If you add a tag to the topline of a test in the feature file([example](https://github.com/ii/xds-test-harness/blob/update-gcp/features/subscriptions.feature#L125)), 
you can run the harness for just this tag with the -t flag:

``` sh
go run -t "@mytest"
```

# Design

The suite is made of tests, a test runner, and an adapter api that target
servers can implement to work with the runner.

The adapter is meant to be integrated into the target, and so this
repo only holds the API spec. The full design can be read in our [design
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
  which implements our features into go code.
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

To learn more about the xDS protocol, you can read [its documentation on
envoyproxy.io](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)
or through the essay ["The Universal Data Plane
API"](https://blog.envoyproxy.io/the-universal-data-plane-api-d15cec7a)

The core testing functionality is handled by the excellent [godog
library](https://github.com/cucumber/godog)

The suite is built with and is testing gRPC services.
[grpc.io](https://grpc.io/) has a great [introduction to
grpc](https://grpc.io/docs/what-is-grpc/introduction/) and [tutorial for
go](https://grpc.io/docs/languages/go/basics/)
