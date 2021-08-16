Feature: Fetching Resources
  For each guy i should be able to do stuff

  Scenario: A wildcard CDS request should return all cluster resources
    Given a target setup with the following "resources":
    ```
    clusters:
    - name: A
      version: 1
    ```
    When the Runner sends a CDS wildcard request
    Then the Runner receives the following "clusters":
    ```
    - name: A
      version: 1
    ```
