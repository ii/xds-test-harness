package parser

import (
	"fmt"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/ptypes"
	pb "github.com/zachmandeville/tester-prototype/api/adapter"
	"gopkg.in/yaml.v2"
)

func ParseInitConfig(yml []byte) (*InitConfig, error) {
	var initConfig InitConfig
	err := yaml.Unmarshal(yml, &initConfig)
	if err != nil {
		return nil, err
	}
	return &initConfig, err
}

func YamlToSnapshot(yml string) (*pb.Snapshot, error) {
	var s Snapshot
	err := yaml.Unmarshal([]byte(yml), &s)
	if err != nil {
		return nil, err
	}

	snapshot := &pb.Snapshot{
		Node:    s.Node,
		Version: s.Version,
	}
	if s.Resources.Endpoints != nil {
		endpoints := &pb.Endpoints{}
		for _, e := range s.Resources.Endpoints {
			endpoints.Items = append(endpoints.Items, &pb.Endpoints_Endpoint{
				Name: e.Name,
			})
		}
	}
	if s.Resources.Clusters != nil {
		clusters := &pb.Clusters{}
		for _, c := range s.Resources.Clusters {
			clusters.Items = append(clusters.Items, &pb.Clusters_Cluster{
				Name:           c.Name,
				ConnectTimeout: map[string]int32{"seconds": int32(c.ConnectTimeout.Seconds)},
			})
		}
		snapshot.Clusters = clusters
	}
	if s.Resources.Routes != nil {
		routes := &pb.Routes{}
		for _, r := range s.Resources.Routes {
			routes.Items = append(routes.Items, &pb.Routes_Route{
				Name: r.Name,
			})
		}
	}
	if s.Resources.Routes != nil {
		routes := &pb.Routes{}
		for _, r := range s.Resources.Routes {
			routes.Items = append(routes.Items, &pb.Routes_Route{
				Name: r.Name,
			})
		}
	}
	if s.Resources.Listeners != nil {
		listener := &pb.Listeners{}
		for _, l := range s.Resources.Listeners {
			listener.Items = append(listener.Items, &pb.Listeners_Listener{
				Name: l.Name,
			})
		}
	}
	if s.Resources.Runtimes != nil {
		runtime := &pb.Runtimes{}
		for _, r := range s.Resources.Runtimes {
			runtime.Items = append(runtime.Items, &pb.Runtimes_Runtime{
				Name: r.Name,
			})
		}
	}
	if s.Resources.Secrets != nil {
		secret := &pb.Secrets{}
		for _, s := range s.Resources.Secrets {
			secret.Items = append(secret.Items, &pb.Secrets_Secret{
				Name: s.Name,
			})
		}
	}
	return snapshot, nil
}

func NewDiscoveryRequest() TestDiscoveryRequest {
	dr := TestDiscoveryRequest{}
	dr.VersionInfo = ""
	dr.ResponseNonce = ""
	dr.TypeURL = ""
	dr.VersionInfo = ""
	dr.Node.ID = ""
	return dr
}

func NewDiscoveryResponse() DiscoveryResponse {
	dr := DiscoveryResponse{
		VersionInfo: "",
		TypeURL:     "",
		Resources:   []Cluster{},
	}
	return dr
}

func ParseDiscoveryRequest(yml string) (*TestDiscoveryRequest, error) {
	request := NewDiscoveryRequest()
	err := yaml.Unmarshal([]byte(yml), &request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func ParseDiscoveryResponse(dr *envoy_service_discovery_v3.DiscoveryResponse) (*DiscoveryResponse, error) {
	response := DiscoveryResponse{
		VersionInfo: dr.VersionInfo,
		TypeURL:     dr.TypeUrl,
	}
	for i, _ := range dr.GetResources() {
		var cpb cluster.Cluster
		err := ptypes.UnmarshalAny(dr.GetResources()[i], &cpb)
		if err != nil {
			fmt.Printf("anypb error: %v\n", err)
		}
		// TODO: why is this necessary?
		cc := Cluster{
			Name: cpb.Name,
			ConnectTimeout: ConnectTimeout{
				Seconds: cpb.ConnectTimeout.Seconds,
			},
		}

		response.Resources = append(response.Resources, cc)
	}
	return &response, nil
}
