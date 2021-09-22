@old
Feature: Fetching Resources
  Client can subscribe and unsubcribe from resources, and receive updates when
  any subscribed resources change.
  # these features come from this list of test cases
  # => https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E/edit#

  # test case 1 in list
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

  # test case 2
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
    When the Client subscribes to wildcard CDS
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 1
    resources:
      clusters:
      - name: D
      - name: E
      - name: F
    ```
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

  # test case 3
  Scenario: A client subscribed to a wildcard CDS should receive response when new resources are created
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: A
      - name: B
    ```
    When the Client subscribes to wildcard CDS
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 1
    resources:
      clusters:
      - name: A
      - name: B
    ```
    When the Target is updated to the following state:
    ```
    version: 2
    resources:
      clusters:
      - name: A
      - name: B
      - name: C
    ```

    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 2
    resources:
      clusters:
      - name: A
      - name: B
      - name: C
    ```
    And the Client sends an ACK to which the server does not respond

  # test case 4
  @wip
  Scenario:  When subscribing to specific CDS resources, receive only these resources
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: A
      - name: B
    ```
    When the Client subscribes to the following resources:
    ```
    - A
    - B
    ```
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 1
    resources:
      clusters:
      - name: A
      - name: B
    ```
    And the Client sends an ACK to which the server does not respond

  # test case 5
  Scenario: When subscribing to specific resources, receive response when those resources change
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: X
      - name: Y
    ```
    When the Client subscribes to the following resources:
    ```
    - X
    - Y
    ```
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 1
    resources:
      clusters:
      - name: X
      - name: Y
    ```
    When the Target is updated to the following state:
    ```
    version: 2
    resources:
      clusters:
      - name: X
      - name: Y
    ```
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 2
    resources:
      clusters:
      - name: X
      - name: Y
    ```
    And the Client sends an ACK to which the server does not respond

  # test case 6
  Scenario: When subscribing to resources that don't exist, receive response when they are created
    Given a target setup with the following state:
    ```
    version: 1
    resources:
      clusters:
      - name: A
    ```
    When the Client subscribes to the following resources:
    ```
    - A
    - B
    ```
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 1
    resources:
      clusters:
      - name: A
    ```
    When the Target is updated to the following state:
    ```
    version: 2
    resources:
      clusters:
      - name: A
      - name: B
    ```
    Then the Client receives the following version and clusters, along with a nonce:
    ```
    version: 2
    resources:
      clusters:
      - name: A
      - name: B
    ```
    And the Client sends an ACK to which the server does not respond

