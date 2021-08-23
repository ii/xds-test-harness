Feature: Fetching Resources
  For each guy i should be able to do stuff

  Scenario: A wildcard CDS request should return all cluster resources
    Given a target setup with the following state:
    ```
    node: 'test-id'
    version: 1
    resources:
      clusters:
      - name: A
    ```
    When the Runner sends its first CDS wildcard request to "test-id"
    Then the Runner receives the following clusters:
    ```
    version: 1
    resources:
      clusters:
      - name: A
    ```

  Scenario: When a subscribed resource is updated, the update should be sent to the client
    Given a target setup with the following state:
    ```
    node: 'test-id'
    version: 1
    resources:
      clusters:
      - name: A
    ```
    When cluster "A" is updated to version "2" after Runner subscribed to CDS
    Then the Runner receives the following version and clusters:
    ```
    version: 2
    resources:
      clusters:
      - name: A
    ```
