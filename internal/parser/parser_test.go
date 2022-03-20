package parser

import (
	"testing"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	pb "github.com/ii/xds-test-harness/api/adapter"
	"github.com/ii/xds-test-harness/internal/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	testResources = []string{"A", "B", "C"}
)

func TestToEndpoints(t *testing.T) {
	expected := &pb.Endpoints{
		Items: []*pb.Endpoint{
			{
				Name:    "A",
				Cluster: "A",
			},
			{
				Name:    "B",
				Cluster: "B",
			},
			{
				Name:    "C",
				Cluster: "C",
			},
		},
	}

	actual := ToEndpoints(testResources)

	for i, v := range actual.Items {
		exp := expected.Items[i]
		if v.Name != exp.Name {
			t.Errorf("Expected endpoint name doesn't match actual: %v, %v", v.Name, exp.Name)
		}
		if v.Cluster != exp.Cluster {
			t.Errorf("Expected endpoint cluster doesn't match actual: %v, %v", v.Cluster, exp.Cluster)
		}
	}
}

func TestToClusters(t *testing.T) {
	expected := &pb.Clusters{
		Items: []*pb.Cluster{
			{
				Name:           "A",
				ConnectTimeout: map[string]int32{"seconds": 5},
			},
			{
				Name:           "B",
				ConnectTimeout: map[string]int32{"seconds": 5},
			},
			{
				Name:           "C",
				ConnectTimeout: map[string]int32{"seconds": 5},
			},
		},
	}

	actual := ToClusters(testResources)
	for i, v := range actual.Items {
		exp := expected.Items[i]
		if v.Name != exp.Name {
			t.Errorf("Expected cluster name doesn't match actual: %v, %v", v.Name, exp.Name)
		}
		if v.ConnectTimeout["seconds"] != exp.ConnectTimeout["seconds"] {
			t.Errorf("Expected cluster timeout doesn't match actual: %v, %v", v.ConnectTimeout, exp.ConnectTimeout)
		}
	}
}

func TestToRoutes(t *testing.T) {
	expected := &pb.Routes{
		Items: []*pb.Route{
			{
				Name: "A",
			},
			{
				Name: "B",
			},
			{
				Name: "C",
			},
		},
	}

	actual := ToRoutes(testResources)

	for i, v := range actual.Items {
		exp := expected.Items[i]
		if v.Name != exp.Name {
			t.Errorf("Expected route name doesn't match actual: %v, %v", v.Name, exp.Name)
		}
	}
}

func TestToListeners(t *testing.T) {
	expected := &pb.Listeners{
		Items: []*pb.Listener{
			{
				Name: "A",
			},
			{
				Name: "B",
			},
			{
				Name: "C",
			},
		},
	}

	actual := ToListeners(testResources)

	for i, v := range actual.Items {
		exp := expected.Items[i]
		if v.Name != exp.Name {
			t.Errorf("Expected listener name doesn't match actual: %v, %v", v.Name, exp.Name)
		}
	}
}

func TestToRuntimes(t *testing.T) {
	expected := &pb.Runtimes{
		Items: []*pb.Runtime{
			{
				Name: "A",
			},
			{
				Name: "B",
			},
			{
				Name: "C",
			},
		},
	}

	actual := ToRuntimes(testResources)

	for i, v := range actual.Items {
		exp := expected.Items[i]
		if v.Name != exp.Name {
			t.Errorf("Expected runtime name doesn't match actual: %v, %v", v.Name, exp.Name)
		}
	}
}

func TestToSecrets(t *testing.T) {
	expected := &pb.Secrets{
		Items: []*pb.Secret{
			{
				Name: "A",
			},
			{
				Name: "B",
			},
			{
				Name: "C",
			},
		},
	}

	actual := ToSecrets(testResources)

	for i, v := range actual.Items {
		exp := expected.Items[i]
		if v.Name != exp.Name {
			t.Errorf("Expected secret name doesn't match actual: %v, %v", v.Name, exp.Name)
		}
	}
}

func TestServiceToTypeURL(t *testing.T) {
	yah := "lds"
	yah2 := "CdS"
	nah := "zds"

	if _, v := ServiceToTypeURL(yah); v != TypeUrlLDS {
		t.Errorf("Incorrect service given back(expected, actual): %v %v", TypeUrlLDS, v)
	}

	if _, v := ServiceToTypeURL(yah2); v != TypeUrlCDS {
		t.Errorf("Incorrect service given back(expected, actual): %v %v", TypeUrlLDS, v)
	}
	if err, v := ServiceToTypeURL(nah); err == nil {
		t.Errorf("Unknown type urls should return err. Instead received %v", v)
	}
}

func TestResourceNames(t *testing.T) {
	names := []string{"tui", "kea", "kakapo"}

	// Test Listener Resources
	listeners := []*anypb.Any{}
	for _, name := range names {
		dst := &anypb.Any{}
		src := &listener.Listener{Name: name}
		opts := proto.MarshalOptions{}
		err := anypb.MarshalFrom(dst, src, opts)
		if err != nil {
			t.Errorf("Error marshalling listener to anypb.any: %v", err)
		}
		listeners = append(listeners, dst)
	}
	ldsResponse := &envoy_service_discovery_v3.DiscoveryResponse{
		VersionInfo: "1",
		Resources:   listeners,
		TypeUrl:     TypeUrlLDS,
		Nonce:       "1",
	}

	ldsNames, err := ResourceNames(ldsResponse)
	if err != nil {
		t.Errorf("Error getting Resource names, when not expecting error.\nerr:%v", err)
	}

	for _, name := range names {
		inResourceNames := itemInSlice(name, ldsNames)
		if !inResourceNames {
			t.Errorf("Could not find required resource name in parsed resource names.\nname: %v\nresources: %v", name, ldsNames)
		}
	}

	// Test Cluster Resources
	clusters := []*anypb.Any{}
	for _, name := range names {
		dst := &anypb.Any{}
		src := &cluster.Cluster{Name: name}
		opts := proto.MarshalOptions{}
		err := anypb.MarshalFrom(dst, src, opts)
		if err != nil {
			t.Errorf("Error marshalling cluster to anypb.any: %v", err)
		}
		clusters = append(clusters, dst)
	}

	cdsResponse := &envoy_service_discovery_v3.DiscoveryResponse{
		VersionInfo: "1",
		Resources:   clusters,
		TypeUrl:     TypeUrlCDS,
		Nonce:       "1",
	}

	cdsNames, err := ResourceNames(cdsResponse)
	if err != nil {
		t.Errorf("Error getting Resource names, when not expecting error.\nerr:%v", err)
	}

	for _, name := range names {
		inResourceNames := itemInSlice(name, cdsNames)
		if !inResourceNames {
			t.Errorf("Could not find required cds name in parsed resource names.\nname: %v\nresources: %v", name, cdsNames)
		}
	}

	// Test Endpoint Resources
	endpoints := []*anypb.Any{}
	for _, name := range names {
		dst := &anypb.Any{}
		src := &endpoint.ClusterLoadAssignment{ClusterName: name}
		opts := proto.MarshalOptions{}
		err := anypb.MarshalFrom(dst, src, opts)
		if err != nil {
			t.Errorf("Error marshalling endpoint to anypb.any: %v", err)
		}
		endpoints = append(endpoints, dst)
	}

	edsResponse := &envoy_service_discovery_v3.DiscoveryResponse{
		VersionInfo: "1",
		Resources:   endpoints,
		TypeUrl:     TypeUrlEDS,
		Nonce:       "1",
	}

	edsNames, err := ResourceNames(edsResponse)
	if err != nil {
		t.Errorf("Error getting Resource names, when not expecting error.\nerr:%v", err)
	}

	for _, name := range names {
		inResourceNames := itemInSlice(name, edsNames)
		if !inResourceNames {
			t.Errorf("Could not find required eds name in parsed resource names.\nname: %v\nresources: %v", name, edsNames)
		}
	}

	// Test Route Resources
	routes := []*anypb.Any{}
	for _, name := range names {
		dst := &anypb.Any{}
		src := &route.RouteConfiguration{Name: name}
		opts := proto.MarshalOptions{}
		err := anypb.MarshalFrom(dst, src, opts)
		if err != nil {
			t.Errorf("Error marshalling route to anypb.any: %v", err)
		}
		routes = append(routes, dst)
	}

	rdsResponse := &envoy_service_discovery_v3.DiscoveryResponse{
		VersionInfo: "1",
		Resources:   routes,
		TypeUrl:     TypeUrlRDS,
		Nonce:       "1",
	}

	rdsNames, err := ResourceNames(rdsResponse)
	if err != nil {
		t.Errorf("Error getting Resource names, when not expecting error.\nerr:%v", err)
	}

	for _, name := range names {
		inResourceNames := itemInSlice(name, rdsNames)
		if !inResourceNames {
			t.Errorf("Could not find required rds name in parsed resource names.\nname: %v\nresources: %v", name, rdsNames)
		}
	}
}

func itemInSlice(item string, slice []string) bool {
	for _, sliceItem := range slice {
		if item == sliceItem {
			return true
		}
	}
	return false
}

func TestParseSupportedVariants(t *testing.T) {
	yah := []string{"sotw Non-Aggregated", "incremental aggregated"}
	yahTypes := []types.Variant{types.SotwNonAggregated, types.IncrementalAggregated}

	nah := []string{"kakapo", "kea", "tui"}

	err, yahVars := ParseSupportedVariants(yah)
	if err != nil {
		t.Errorf("Error parsing variants when expecting no err: %v", err)
	}
	for i, v := range yahVars {
		if v != yahTypes[i] {
			t.Errorf("Parsed Variant does not match expected. expected:%v, actual:%v", yahTypes[i], v)
		}
	}

	err, _ = ParseSupportedVariants(nah)
	if err == nil {
		t.Errorf("Parsing should return error when given bad variant strings. It did not.")
	}
}

func TestValuesFromConfig(t *testing.T) {
	config := "../../testdata/config.yaml"
	expected := map[string]string{
		"nodeID":  "testaroo",
		"target":  "12000",
		"adapter": "13000",
	}
	expectedVariants := []types.Variant{types.SotwNonAggregated, types.IncrementalAggregated}
	target, adapter, nodeID, variants := ValuesFromConfig(config)
	if target != expected["target"] {
		t.Errorf("Target not parsed from config properly. expected: %v actual: %v", expected["target"], target)
	}
	if adapter != expected["adapter"] {
		t.Errorf("Adapter not parsed from config properly. expected: %v actual: %v", expected["adapter"], adapter)
	}
	if nodeID != expected["nodeID"] {
		t.Errorf("NodeID not parsed from config properly. expected: %v actual: %v", expected["nodeID"], nodeID)
	}
	for i, variant := range variants {
		if variant != expectedVariants[i] {
			t.Errorf("Variant not parsed correctly. expected: %v, actual: %v", expectedVariants[i], variant)
		}
	}
}
