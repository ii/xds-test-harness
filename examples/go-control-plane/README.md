# Example xDS Server: go-control-plane

This is an example of a trivial xDS V3 control plane server, that uses [envoy's go control plane](https://github.com/envoyproxy/go-control-plane/).  It implements the xds-test-harness adapter for setting and updating the server state.  It is not meant as an actual envoy control plane, and maintains just a simple snapshot state.

To run it, from the root of this repo, invoke:

```
go run examples/go-control-plane/main.go
```
Once it is running, you can run the test suite in a separate window with:
```
go run .
```

## Files

* [main/main.go](main/main.go) is the example program entrypoint.  It instantiates the cache and xDS server and runs the xDS server process.
* [adapter.go](adapter.go) implementation of the [adapter api](https://github.com/ii/xds-test-harness/blob/main/api/adapter/adapter.proto).
* [server.go](server.go) runs the xDS control plane server.
* [logger.go](logger.go) implements the `pkg/log/Logger` interface which provides logging services to the cache.
