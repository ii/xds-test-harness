Feature: Subscribing to Resources
  Client can do wildcard subscriptions or normal subscriptions
  and receive updates when any subscribed resources change.

  These features come from this list of test cases:
  https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E

  @sotw @non-aggregated @aggregated
  Scenario Outline: The service should send all resources on a wildcard request.
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the Client receives the <expected resources> and <starting version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources | expected resources |
      | "CDS"   | "1"              | "A,B,C"   | "C,A,B"            |
      | "LDS"   | "1"              | "D,E,F"   | "F,D,E"            |


  @sotw @non-aggregated @aggregated
  Scenario Outline: The service should send updates to the client
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the Client receives the <expected resources> and <starting version>
    When a <chosen resource> of the <service> is updated to the <next version>
    Then the Client receives the <expected resources> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | next version | resources | expected resources | chosen resource |
      | "CDS"   | "1"              | "2"          | "A,B,C"   | "A,B,C"            | "A"             |
      | "LDS"   | "1"              | "2"          | "D,E,F"   | "D,E,F"            | "D"             |


  @sotw @non-aggregated @aggregated
  Scenario Outline: Wildcard subscriptions receive updates when new resources are added
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the Client receives the <resources> and <starting version>
    When a <new resource> is added to the <service> with <next version>
    Then the Client receives the <expected resources> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources | new resource | expected resources | next version |
      | "CDS"   | "1"              | "A,B,C"   | "D"          | "A,B,C,D"          | "2"          |
      | "LDS"   | "1"              | "D,E,F"   | "G"          | "D,E,F,G"          | "2"          |


  @sotw @non-aggregated @aggregated
  Scenario:  When subscribing to specific resources, receive only these resources
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <subset of resources> and <starting version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources | subset of resources |
      | "CDS"   | "1"              | "A,B,C,D" | "B,D"               |
      | "LDS"   | "1"              | "G,B,L,D" | "L,G"               |
      | "RDS"   | "1"              | "A,B"     | "A,B"               |
      | "EDS"   | "1"              | "A,B"     | "A,B"               |


  @CDS @LDS @sotw @non-aggregated @aggregated
  Scenario: When subscribing to specific resources, receive response when those resources change
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <subset of resources> and <starting version>
    When a <subscribed resource> of the <service> is updated to the <next version>
    Then the Client receives the <subset of resources> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | subscribed resource | next version |
      | "CDS"   | "1"              | "A,B,C,D"   | "B,D"               | "B"                 | "2"          |
      | "LDS"   | "1"              | "G,B,L,D"   | "L,G"               | "G"                 | "2"          |


  @sotw @non-aggregated @aggregated
  Scenario: When subscribing to specific resources, receive response when those resources change
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <subset of resources> and <starting version>
    When a <subscribed resource> of the <service> is updated to the <next version>
    Then the Client receives the <subscribed resource> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | subscribed resource | next version |
      | "RDS"   | "1"              | "A,B,C,D"   | "B,D"               | "B"                 | "2"          |
      | "EDS"   | "1"              | "A,B,C,D"   | "B,D"               | "B"                 | "2"          |


  @CDS @LDS @sotw @non-aggregated @aggregated
  Scenario: When subscribing to resources that don't exist, receive response when they are created
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <existing subset> and <starting version>
    When a <chosen resource> is added to the <service> with <next version>
    Then the Client receives the <subset of resources> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | existing subset | chosen resource | next version |
      | "CDS"   | "1"              | "A,B,C,D"   | "A,Z"               | "A"             | "Z"             | "2"          |
      | "LDS"   | "1"              | "G,B,L,D"   | "B,D,X"             | "B,D"           | "X"             | "2"          |


  @sotw @non-aggregated @aggregated
  Scenario: When subscribing to resources that don't exist, receive response when they are created
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <existing subset> and <starting version>
    When a <chosen resource> is added to the <service> with <next version>
    Then the Client receives the <chosen resource> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | existing subset | chosen resource | next version |
      | "RDS"   | "1"              | "A,B,C,D"   | "A,Z"               | "A"             | "Z"             | "2"          |
      | "EDS"   | "1"              | "A,B,C,D"   | "A,Z"               | "A"             | "Z"             | "2"          |