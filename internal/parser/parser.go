package parser

import (
	"fmt"
	"math/rand"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/ptypes"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"gopkg.in/yaml.v2"
)

func randomAddress() string {
	var (
		consonants = []rune("bcdfklmnprstwyz")
		vowels     = []rune("aou")
		tld = []string{".biz",".com",".net",".org"}

		domain     = ""
	)
	rand.Seed(time.Now().UnixNano())
	length := 6 + rand.Intn(12)

	for i := 0; i < length; i++ {
		consonant := string(consonants[rand.Intn(len(consonants))])
		vowel := string(vowels[rand.Intn(len(vowels))])

		domain = domain + consonant + vowel
	}
	return domain + tld[rand.Intn(len(tld))]
}


func YamlToDesiredState(yml string) (*Snapshot, error) {
	var s *Snapshot
	err := yaml.Unmarshal([]byte(yml), &s)
	return s, err
}

func ParseResourceList(resources string)  ([]string, error) {
	var resourceList []string
	err := yaml.Unmarshal([]byte(resources), &resourceList)
	return resourceList, err
}

func YamlToSnapshot(nodeID string, yml string) (*pb.Snapshot, error) {
	s, err := YamlToDesiredState(yml)
	if err != nil {
		return nil, err
	}

	snapshot := &pb.Snapshot{
		Node:    nodeID,
		Version: s.Version,
	}
	if s.Resources.Endpoints != nil {
		endpoints := &pb.Endpoints{}
		for _, e := range s.Resources.Endpoints {
			endpoints.Items = append(endpoints.Items, &pb.Endpoints_Endpoint{
				Name: e.Name,
			})
		}
		snapshot.Endpoints = endpoints
	}

	if s.Resources.Clusters != nil {
		clusters := &pb.Clusters{}
		for _, c := range s.Resources.Clusters {
			clusters.Items = append(clusters.Items, &pb.Clusters_Cluster{
				Name:           c.Name,
				ConnectTimeout: map[string]int32{"seconds": 5},
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

	if s.Resources.Listeners != nil {
		fmt.Println("snap listeners", s.Resources.Listeners)
		listeners := &pb.Listeners{}
		for _, l := range s.Resources.Listeners {
			listeners.Items = append(listeners.Items, &pb.Listeners_Listener{
				Name: l.Name,
				Address: randomAddress(),
			})
		}
		snapshot.Listeners = listeners
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

func ParseDiscoveryResponse(dr *envoy_service_discovery_v3.DiscoveryResponse) (*DiscoveryResponse, error) {
	response := DiscoveryResponse{
		VersionInfo: dr.GetVersionInfo(),
		TypeURL:     dr.GetTypeUrl(),
		Resources:   []Cluster{},
		Nonce:       dr.GetNonce(),
	}
	for i := range dr.GetResources() {
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
