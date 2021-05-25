Feature: Conformance ACK/NACK
  Discovery Requests and Responses should follow the behaviour outlined in the
  API docs.

  Background:
    Given "adapter" is reachable via grpc
    And "target" is reachable via grpc
    And "target_adapter" is reachable via grpc

  Scenario:
    Given a Target setup with snapshot matching yaml:
     ```
     ---
     Node:
       name: test-id
     Resources:
     - Version: '1'
       Items: {}
     - Version: '1'
       Items:
         foo:
           Resource:
             name: foo
             connect_timeout:
               seconds: 5
     - Version: '1'
       Items: {}
     - Version: '1'
       Items: {}
     - Version: '1'
       Items: {}
     - Version: '1'
       Items: {}
     - Version: '1'
       Items: {}
     ```
     When I send a discovery request matching yaml:
     ```
     version_info:
     node: { id: test-id }
     resource_names:
     type_url: type.googleapis.com/envoy.config.cluster.v3.Cluster
     response_nonce:
     ```
     Then I get a discovery response matching json:
     ```
     {
       "versionInfo":"1",
       "resources":[
         {"typeUrl":"type.googleapis.com/envoy.config.cluster.v3.Cluster",
          "value":"CgNmb28iAggF"}
        ],
        "typeUrl":"type.googleapis.com/envoy.config.cluster.v3.Cluster"
     }
     ```
