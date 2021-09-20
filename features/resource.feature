Feature: Fetching Resources
  Client can subscribe and unsubcribe from resources, and receive updates when
  any subscribed resources change.

  Scenario: Server should send all CDS resources on a CDS wildcard request.
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: A
      - name: B
      - name: C
    ```
    When the Client subscribes to wildcard CDS
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 1
    resources:
      clusters:
      - name: A
      - name: B
      - name: C
    ```
    And the Client sends an ACK to which the server does not respond

  Scenario: When a subscribed resource is updated, the update should be sent to the client
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: D
      - name: E
      - name: F
    ```
    And the Client subscribes to wildcard CDS
    When the Target is updated to the following state:
    ```
    version: 2
    resources:
      clusters:
      - name: D
      - name: E
      - name: F
    ```

    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 2
    resources:
      clusters:
      - name: D
      - name: E
      - name: F
    ```
    And the Client sends an ACK to which the server does not respond
