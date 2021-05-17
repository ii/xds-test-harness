Feature: Conformance ACK/NACK
  Discovery Requests and Responses should follow the behaviour outlined in the
  API docs.

  Background:
    Given an Adapter located at "localhost:6767"
    And a Target located at "localhost:18000"
    And a Shim located at "localhost:17000"

  Scenario:
    Given a Target with clusters specified with yaml:
     ```
     name: test_config
     spec:
       clusters:
       - name: foo
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
