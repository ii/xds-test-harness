Feature: A client can subscribe and unsubcribe from an arbitary amount of resources, and receive updates for only the resources they're subscribed to
  Background:
    Given "adapter" is reachable via gRPC
    And "target" is reachable via gRPC

  @delta
  Scenario: I can subscribe to a subset of resources
    Given a target setup with the following resources:
    ```
    clusters:
    - foo
    - bar
    - baz
    - echo
    - fun
    ```
    When I subscribe to CDS for these clusters:
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
