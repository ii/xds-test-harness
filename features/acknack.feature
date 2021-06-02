Feature: Conformance ACK
  Discovery Requests and Responses should follow the behaviour outlined in the
  API docs.

  Background:
    Given "adapter" is reachable via gRPC
    And "target" is reachable via gRPC

  Scenario:
    Given a Target setup with snapshot matching yaml:
     ```
     ---
     node: test-id
     version: "1"
     resources:
       endpoints:
       clusters:
       - name: foo
         connect_timeout:
           seconds: 5
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
