
Feature: Fetching Resources with LDS and CDS
  Client can do wildcard subscriptions or normal subscriptions
  and receive updates when any subscribed resources change.

  These features come from this list of test cases:
  https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E


  Scenario Outline: The service should send all resources on a wildcard request.
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the Client receives the <expected resources> and <starting version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      # Steps 3 and 5 should fail
      | service | starting version | resources | expected resources |
      | "CDS"   | "1"              | "A,B,C"   | "A,B,C"            |
      | "CDS"   | "1"              | "A,B,C"   | "C,A,B"            |
      | "CDS"   | "1"              | "A,B,C"   | "B,A,D"            |
      | "LDS"   | "1"              | "D,E,F"   | "D,E,F"            |
      | "LDS"   | "1"              | "D,E,F"   | "F,A,I,L"          |
      | "LDS"   | "1"              | "D,E,F"   | "F,D,E"            |



  Scenario Outline: The service should send updates to the client
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the client receives the <expected resources> and <starting version> for <service>
    When a <chosen resource> of the <service> is updated to the <next version>
    Then the Client receives the <expected resources> and <next version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | next version | resources | expected resources | chosen resource |
      | "CDS"   | "1"              | "2"          | "A,B,C"   | "A,B,C"            | "A"             |
      | "CDS"   | "1"              | "2"          | "A,B,C"   | "C,A,B"            | "A"             |
      | "CDS"   | "1"              | "2"          | "A,B,C"   | "B,C,A"            | "A"             |
      | "LDS"   | "1"              | "2"          | "D,E,F"   | "D,E,F"            | "D"             |
      | "LDS"   | "1"              | "2"          | "D,E,F"   | "D,E,F"            | "D"             |
      | "LDS"   | "1"              | "2"          | "D,E,F"   | "F,D,E"            | "D"             |



  Scenario Outline: Wildcard subscriptions receive updates when new resources are added
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the client receives the <resources> and <starting version> for <service>
    When a <new resource> is added to the <service> with <next version>
    Then the Client receives the <expected resources> and <next version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources | new resource | expected resources | next version |
      | "CDS"   | "1"              | "A,B,C"   | "D"          | "A,B,C,D"          | "2"          |
      | "CDS"   | "1"              | "A,B,C"   | "E"          | "A,B,C,E"          | "2"          |
      | "CDS"   | "1"              | "A,B,C"   | "F"          | "A,B,C,F"          | "2"          |
      | "LDS"   | "1"              | "D,E,F"   | "G"          | "D,E,F,G"          | "2"          |
      | "LDS"   | "1"              | "D,E,F"   | "H"          | "D,E,F,H"          | "2"          |
      | "LDS"   | "1"              | "D,E,F"   | "I"          | "D,E,F,I"          | "2"          |



  Scenario:  When subscribing to specific CDS resources, receive only these resources
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the client receives the <subset of resources> and <starting version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      # Test 5 should fail
      | service | starting version | resources   | subset of resources |
      | "CDS"   | "1"              | "A,B,C,D"   | "B,D"               |
      | "CDS"   | "1"              | "B,C,A,"    | "C"                 |
      | "CDS"   | "1"              | "F,A,B,C,D" | "A,C,D"             |
      | "LDS"   | "1"              | "G,B,L,D"   | "B,D"               |
      | "LDS"   | "1"              | "B,L,G,"    | "L, H"              |
      | "LDS"   | "1"              | "F,G,B,L,D" | "G,L,D"             |



  Scenario: When subscribing to specific resources, receive response when those resources change
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the client receives the <subset of resources> and <starting version> for <service>
    When a <subscribed resource> of the <service> is updated to the <next version>
    Then the Client receives the <subset of resources> and <next version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | subscribed resource | next version |
      | "CDS"   | "1"              | "A,B,C,D"   | "B,D"               | "B"                 | "2"          |
      | "CDS"   | "1"              | "B,C,A,"    | "C"                 | "C"                 | "2"          |
      | "CDS"   | "1"              | "F,A,B,C,D" | "A,C,D"             | "C"                 | "2"          |
      | "LDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                 | "2"          |
      | "LDS"   | "1"              | "B,L,G,"    | "L,G"               | "G"                 | "2"          |
      | "LDS"   | "1"              | "F,G,B,L,D" | "G,L,D"             | "D"                 | "2"          |



  Scenario: When subscribing to resources that don't exist, receive response when they are created
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the client receives the <existing subset> and <starting version> for <service>
    When a <chosen resource> is added to the <service> with <next version>
    Then the Client receives the <subset of resources> and <next version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | existing subset | chosen resource | next version |
      | "CDS"   | "1"              | "A,B,C,D"   | "A,Z"               | "A"             | "Z"             | "2"          |
      | "CDS"   | "1"              | "B,C,A,"    | "C,Y"               | "C"             | "Y"             | "2"          |
      | "CDS"   | "1"              | "F,A,B,C,D" | "A,C,D,X"           | "A,C,D"         | "X"             | "2"          |
      | "LDS"   | "1"              | "G,B,L,D"   | "B,D,X"             | "B,D"           | "X"             | "2"          |
      | "LDS"   | "1"              | "B,L,G,"    | "L,G,Y"             | "G,L"           | "Y"             | "2"          |
      | "LDS"   | "1"              | "F,G,B,L,D" | "G,L,D,Z"           | "D,G,L"         | "Z"             | "2"          |


  @wip
  Scenario: Client can unsubcribe from some resources
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the client receives the <subset of resources> and <starting version> for <service>
    When the Client updates subscription to a <resource from subset> of <service> with <starting version>
    And a <resource from subset> of the <service> is updated to the <next version>
    Then the Client receives the <resource from subset> and <next version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | resource from subset |   next version  |
      | "CDS"   | "1"              | "A,B,C,D"   | "A,B"               | "A"                  |   "2"           |
      | "CDS"   | "1"              | "F,A,B,C,D" | "C,A,B"             | "A,C"                |   "2"           |
      | "LDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                  |   "2"           |
      | "LDS"   | "1"              | "B,L,A,G,"  | "L,G,B"             | "L,G"                |   "2"           |
