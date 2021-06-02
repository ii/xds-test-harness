package xdscache

import "google.golang.org/protobuf/types/known/durationpb"

type Listener struct {
	Name       string
	Address    string
	Port       uint32
	RouteNames []string
}

type Route struct {
	Name    string
	Prefix  string
	Cluster string
}

type Cluster struct {
	Name           string
	ConnectTimeout *durationpb.Duration
	Endpoints      []Endpoint
}

type Endpoint struct {
	UpstreamHost string
	UpstreamPort uint32
}
