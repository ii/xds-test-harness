Feature: Fetching Resources
  Client can subscribe and unsubcribe from resources, and receive updates when
  any subscribed resources change.

  Scenario: A wildcard CDS request should return all cluster resources
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: A
    ```
    When the Client sends an initial CDS wildcard request
    Then the Client receives the following version and clusters:
    ```
    version: 1
    resources:
      clusters:
      - name: A
    ```

  Scenario: When a subscribed resource is updated, the update should be sent to the client
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: A
    ```
    When cluster "A" is updated to version "2" after Client subscribed to CDS
    Then the Client receives the following version and clusters:
    ```
    version: 2
    resources:
      clusters:
      - name: A
    ```
