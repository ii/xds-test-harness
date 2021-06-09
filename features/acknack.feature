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
    Then I get a discovery response matching yaml:
    ```
    version_info: "1"
    resources:
      - name: foo
        connect_timeout:
          seconds: 5
    type_url: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
    ```

  Scenario:
    Given a Target setup with snapshot matching yaml:
    ```
    ---
    node: test-id
    version: "1"
    resources:
      clusters:
        - name: foo
      connect_timeout:
        seconds: 5
    ```
    And a CDS stream was initiated with a discovery request matching yaml:
    ```
    version_info:
    node: { id: test-id }
    resource_names:
    type_url: type.googleapis.com/envoy.config.cluster.v3.Cluster
    response_nonce:
    ```
    And the stream was ACKed with a discovery request matching yaml:
    ```
    version_info: "1"
    node: { id: test-id }
    resource_names:
    type_url: type.googleapis.com/envoy.config.cluster.v3.Cluster
    response_nonce: "1"
    ```
    When Target is updated to match yaml:
    ```
    ---
    node: test-id
    version: "2"
    resources:
      clusters:
      - name: foo
        connect_timeout:
          seconds: 10
    ```
    Then the client receives a discovery response matching yaml:
    ```
    version_info: "2"
    resources:
      - name: foo
        connect_timeout:
          seconds: 10
    type_url: "type.googleapis.com/envoy.config.cluster.v3.Cluster"
    ```
