package parser

import (
	"fmt"
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

	snapshot := &pb.Snapshot{}
	if s.Resources.Endpoints == nil {
		fmt.Println("NO ENDPOINTS")
	}
	if s.Resources.Clusters != nil {
		clusters := &pb.Clusters{
			Items: []*pb.Clusters_Cluster{},
		}
		for _, c := range s.Resources.Clusters {
			clusters.Items = append(clusters.Items, &pb.Clusters_Cluster{
				Name:           c.Name,
				ConnectTimeout: map[string]string{"seconds": string(c.ConnectTimeout.seconds)},
			})
		}
		snapshot.Clusters = clusters
	}
	fmt.Printf("snapshot: %v", snapshot)
	return snapshot, nil

}

// func ParseYaml(yml string) (*EnvoyConfig, error) {
// 	var config EnvoyConfig

// 	err := yaml.Unmarshal([]byte(yml), &config)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &config, nil
// }

// func NewDiscoveryRequest() TestDiscoveryRequest {
// 	dr := TestDiscoveryRequest{}
// 	dr.VersionInfo = ""
// 	dr.ResponseNonce = ""
// 	dr.TypeURL = ""
// 	dr.VersionInfo = ""
// 	dr.Node.ID = ""
// 	return dr
// }

// func ParseDiscoveryRequest(yml string) (*TestDiscoveryRequest, error) {
// 	request := NewDiscoveryRequest()
// 	err := yaml.Unmarshal([]byte(yml), &request)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &request, nil
// }
