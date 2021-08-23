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
