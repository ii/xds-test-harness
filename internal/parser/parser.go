package parser

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/ii/xds-test-harness/internal/types"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/rs/zerolog/log"
)

const (
	TypeUrlLDS = "type.googleapis.com/envoy.config.listener.v3.Listener"
	TypeUrlCDS = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	TypeUrlRDS = "type.googleapis.com/envoy.config.route.v3.RouteConfiguration"
	TypeUrlEDS = "type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment"
)

func RandomAddress() string {
	var (
		consonants = []rune("bcdfklmnprstwyz")
		vowels     = []rune("aou")
		tld        = []string{".biz", ".com", ".net", ".org"}

		domain = ""
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

func ToEndpoints(resourceNames []string) *pb.Endpoints {
	endpoints := &pb.Endpoints{}
	for _, name := range resourceNames {
		endpoints.Items = append(endpoints.Items, &pb.Endpoint{
			Name:    name,
			Cluster: name,
			Address: RandomAddress(),
		})
	}
	return endpoints
}

func ToClusters(resourceNames []string) *pb.Clusters {
	clusters := &pb.Clusters{}
	for _, name := range resourceNames {
		clusters.Items = append(clusters.Items, &pb.Cluster{
			Name:           name,
			ConnectTimeout: map[string]int32{"seconds": 5},
		})
	}
	return clusters
}

func ToRoutes(resourceNames []string) *pb.Routes {
	routes := &pb.Routes{}
	for _, name := range resourceNames {
		routes.Items = append(routes.Items, &pb.Route{
			Name: name,
		})
	}
	return routes
}

func ToListeners(resourceNames []string) *pb.Listeners {
	listeners := &pb.Listeners{}
	for _, name := range resourceNames {
		listeners.Items = append(listeners.Items, &pb.Listener{
			Name:    name,
			Address: RandomAddress(),
		})
	}
	return listeners
}

func ToRuntimes(resourceNames []string) *pb.Runtimes {
	runtimes := &pb.Runtimes{}
	for _, name := range resourceNames {
		runtimes.Items = append(runtimes.Items, &pb.Runtime{
			Name: name,
		})
	}
	return runtimes
}

func ToSecrets(resourceNames []string) *pb.Secrets {
	secrets := &pb.Secrets{}
	for _, name := range resourceNames {
		secrets.Items = append(secrets.Items, &pb.Secret{
			Name: name,
		})
	}
	return secrets
}

func ServiceToTypeURL(service string) (typeURL string, err error) {
	typeURLs := map[string]string{
		"lds": TypeUrlLDS,
		"cds": TypeUrlCDS,
		"eds": TypeUrlEDS,
		"rds": TypeUrlRDS,
	}
	service = strings.ToLower(service)

	typeURL, ok := typeURLs[service]
	if !ok {
		err = fmt.Errorf("Cannot find type URL for given service: %v", service)
		return typeURL, err
	}
	return typeURL, nil
}

func ResourceNames(res *envoy_service_discovery_v3.DiscoveryResponse) (resourceNames []string, err error) {
	typeUrl := res.TypeUrl
	switch typeUrl {
	case TypeUrlLDS:
		for _, resource := range res.GetResources() {
			listener := &listener.Listener{}
			if err := resource.UnmarshalTo(listener); err != nil {
				return nil, fmt.Errorf("Could not get resource name from %v. err: %v", resource, err)
			}
			resourceNames = append(resourceNames, listener.Name)
		}
	case TypeUrlCDS:
		for _, resource := range res.GetResources() {
			cluster := &cluster.Cluster{}
			if err := resource.UnmarshalTo(cluster); err != nil {
				return nil, fmt.Errorf("Could not get resource name from %v. err: %v", resource, err)
			}
			resourceNames = append(resourceNames, cluster.Name)
		}
	case TypeUrlEDS:
		for _, resource := range res.GetResources() {
			endpointConfig := &endpoint.ClusterLoadAssignment{}
			if err := resource.UnmarshalTo(endpointConfig); err != nil {
				return nil, fmt.Errorf("Could not get resource name from %v. err: %v", resource, err)
			}
			resourceNames = append(resourceNames, endpointConfig.ClusterName)
		}
	case TypeUrlRDS:
		for _, resource := range res.GetResources() {
			route := &route.RouteConfiguration{}
			if err := resource.UnmarshalTo(route); err != nil {
				return nil, fmt.Errorf("Could not get resource name from %v. err: %v", resource, err)
			}
			resourceNames = append(resourceNames, route.Name)
		}
	}
	return resourceNames, err
}

func DeltaResourceNames(res *envoy_service_discovery_v3.DeltaDiscoveryResponse) (resourceNames []string, err error) {
	for _, resource := range res.GetResources() {
			resourceNames = append(resourceNames, resource.Name)
	}
	return resourceNames, err
}

func ParseSupportedVariants(variants []string) (err error, supported []types.Variant) {
	variantMap := map[string]types.Variant{
		"sotw non-aggregated":        types.SotwNonAggregated,
		"sotw aggregated":            types.SotwAggregated,
		"incremental non-aggregated": types.IncrementalNonAggregated,
		"incremental aggregated":     types.IncrementalAggregated,
	}

	for _, v := range variants {
		variant, ok := variantMap[strings.ToLower(v)]
		if !ok {
			err := fmt.Errorf("Config included unrecognized variant. Please remove it and try again: %v\n", variant)
			return err, nil
		}
		supported = append(supported, variant)
	}
	return nil, supported
}

func ValuesFromConfig(config string) (target string, adapter string, nodeID string, supportedVariants []types.Variant) {
	c, err := yaml.ReadFile(config)
	if err != nil {
		log.Fatal().
			Msgf("Cannot read config: %v", config)
	}
	nodeID, err = c.Get("nodeID")
	if err != nil {
		log.Fatal().
			Msgf("Error reading config file for Node ID: %v\n", err)
	}
	target, err = c.Get("targetAddress")
	if err != nil {
		log.Fatal().
			Msgf("Error reading config file for Target Address: %v\n", config)
	}
	adapter, err = c.Get("adapterAddress")
	if err != nil {
		log.Info().
			Msgf("Cannot get adapter address from config file: %v\n", err)
	}
	v, err := yaml.Child(c.Root, "variants")
	if err != nil {
		log.Fatal().Msgf("Error getting variants from config: %v\n", err)
	}
	variants := []string{}
	varsInYaml, ok := v.(yaml.List)
	if ok {
		for i := 0; i < varsInYaml.Len(); i++ {
			node := varsInYaml.Item(i)
			variant := string(node.(yaml.Scalar))
			variants = append(variants, variant)
		}
	}
	err, supportedVariants = ParseSupportedVariants(variants)
	if err != nil {
		log.Fatal().Msgf("Cannot parse supported variants from config: %v", err)
	}
	return target, adapter, nodeID, supportedVariants
}
