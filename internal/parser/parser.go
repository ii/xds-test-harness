package parser

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	pb "github.com/ii/xds-test-harness/api/adapter"
)

func RandomAddress() string {
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

func ToClusters(resources string) *pb.Clusters {
	resourceNames := strings.Split(resources, ",")
	clusters := &pb.Clusters{}

	for _, name := range resourceNames {
		clusters.Items = append(clusters.Items, &pb.Cluster{
			Name: name,
			ConnectTimeout: map[string]int32{"seconds": 5},
		})
	}
	return clusters
}

func ToListeners (resources string) *pb.Listeners {
	resourceNames := strings.Split(resources, ",")
	listeners := &pb.Listeners{}

	for _, name := range resourceNames {
		listeners.Items = append(listeners.Items, &pb.Listener{
			Name: name,
			Address: RandomAddress(),
		})
	}
	return listeners
}

func ParseDiscoveryResponseV2 (res *envoy_service_discovery_v3.DiscoveryResponse) (*SimpleResponse, error) {
	simpRes := &SimpleResponse{}

	simpRes.Version = res.VersionInfo
	simpRes.Nonce = res.Nonce
	simpRes.Resources = []string{}

	if res.TypeUrl == "type.googleapis.com/envoy.config.listener.v3.Listener" {
		for _, resource := range res.GetResources() {
		    listener := &listener.Listener{}
			if err := resource.UnmarshalTo(listener); err != nil {
				fmt.Printf("ERORROROROR: %v", err)
				return nil, err
			}
			simpRes.Resources = append(simpRes.Resources, listener.Name)
		}
	}
	if res.TypeUrl == "type.googleapis.com/envoy.config.cluster.v3.Cluster" {
		for _, resource := range res.GetResources() {
		    cluster := &cluster.Cluster{}
			if err := resource.UnmarshalTo(cluster); err != nil {
				fmt.Printf("ERORROROROR: %v", err)
				return nil, err
			}
			simpRes.Resources = append(simpRes.Resources, cluster.Name)
		}
	}
	return simpRes, nil
}
