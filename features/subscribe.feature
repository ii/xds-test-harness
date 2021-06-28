Feature: A client can subscribe and unsubcribe from an arbitary amount of resources, and receive updates for only the resources they're subscribed to
  Background:
    Given "adapter" is reachable via gRPC
    And "target" is reachable via gRPC

  @delta
  Scenario: I can subscribe to a subset of resources
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
      - name: bar
        connect_timeout:
          seconds: 5
      - name: baz
        connect_timeout:
          seconds: 5
      - name: echo
        connect_timeout:
          seconds: 5
      - name: fun
        connect_timeout:
          seconds: 5
    ```
    When I subscribe to delta CDS for these clusters:
    ```
    - foo
    - bar
    - baz
    ```
    Then I get a response containing these resources:
    ```
    clusters:
    - foo
    - bar
    - baz
    ```
    And the response does not contain these resources:
    ```
    clusters:
    - echo
    - fun
  
    ```
