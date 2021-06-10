# Tester Prototype

This is an exploratory repo whose code is a strong work-in-progress. It is
intended to build understanding on how to communicate with xDS servers and clients. 

At the moment, it is a client expected to send a discovery request to the [go control plane example server](https://github.com/envoyproxy/go-control-plane/blob/main/internal/example/README.md) and get a discovery response.

## Diary
I am recording my "journey of UNDERSTANDING" in the org directory.  These files give 
more context for what I've written and why!
- [Setup](./docs/setup.org)
- [Initial Request](./org/initial-request.org)
- [Streaming Request](./org/streaming-request.org)
- [Test Cases for the xDS Transport Protocol](./org/test-cases-for-xds-transport.org)
- [Basic Test Harness](./org/basic-harness.org)
- [Intermediate Test Harness](./org/intermediate-harness.org)
- [Intermediate Test Harness, Part Two](./org/intermediate-harness-2.org)
