Feature: Fetching Resources with LDS and CDS
  Client can do wildcard subscriptions or normal subscriptions
  and receive updates when any subscribed resources change.

  These features come from this list of test cases:
  https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E

  @sotw @separate @aggregated
  Scenario Outline: The service should send all resources on a wildcard request.
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the Client receives the <expected resources> and <starting version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources | expected resources |
      | "CDS"   | "1"              | "A,B,C"   | "C,A,B"            |
      | "LDS"   | "1"              | "D,E,F"   | "F,D,E"            |


  @sotw @separate @aggregated
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


  @sotw @separate @aggregated
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


  @sotw @separate @aggregated
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


  @CDS @LDS @sotw @separate @aggregated
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


  @sotw @separate @aggregated
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


  @CDS @LDS @sotw @separate @aggregated
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


  @sotw @separate @aggregated
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


  @CDS @LDS @sotw @separate @aggregated
  Scenario: Client can unsubcribe from some resources
    # This test does not check if the final results are only the subscribed resources
    # it is valid(though not desired) for a server to send more than is requested.
    # So the test will pass if client subscribes to A,B,C, unsubscribes from B,C and gets A,B,C back.
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <subset of resources> and <starting version>
    When the Client updates subscription to a <resource from subset> of <service> with <starting version>
    And a <resource from subset> of the <service> is updated to the <next version>
    Then the Client receives the <resource from subset> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | resource from subset |   next version  |
      | "CDS"   | "1"              | "F,A,B,C,D" | "C,A,B"             | "A,C"                |   "2"           |
      | "LDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                  |   "2"           |


  @sotw @separate @aggregated
  Scenario: Client can unsubcribe from some resources
    # difference from test above is use of the word ONLY in the final THEN step
    # This currently does not pass for go-control-plane
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <subset of resources> and <starting version>
    When the Client updates subscription to a <resource from subset> of <service> with <starting version>
    And a <resource from subset> of the <service> is updated to the <next version>
    Then the Client receives only the <resource from subset> and <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | resource from subset |   next version  |
      | "RDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                  |   "2"           |
      | "EDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                  |   "2"           |


  @sotw @separate @aggregated
  Scenario: Client can unsubscribe from all resources
    # This is not working currently, the unsusbcribe is not registered,
    # neither as an unsubscribe nor a new wildcard request
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client subscribes to a <subset of resources> for <service>
    Then the Client receives the <subset of resources> and <starting version>
    When the Client unsubscribes from all resources for <service>
    And a <subset of resources> of the <service> is updated to the <next version>
    Then the Client does not receive any message from <service>

    Examples:
      | service | starting version | resources   | subset of resources | next version |
      | "CDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |
      | "LDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |
      | "RDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |
      | "EDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |


    @sotw @aggregated
    Scenario: Client can subscribe to multiple services via ADS
      Given a target setup with <service>, <resources>, and <starting version>
      When the Client subscribes to a <subset of resources> for <service>
      Then the Client receives the <subset of resources> and <starting version> for <service>
      When the Client subscribes to a <subset of resources> for <other service>
      Then the Client receives the <subset of resources> and <starting version> for <other service>
      When a <subset of resources> of the <service> is updated to the <next version>
      Then the Client receives the <subset of resources> and <next version> for <service>
      # trying out different language for server not responding to acks
      And the service never responds more than necessary

      Examples:
        | service | other service | starting version | resources | subset of resources | next version |
        | "CDS"   | "LDS"         | "1"              | "A,B,C"   | "B"                 | "2"          |
        | "RDS"   | "EDS"         | "1"              | "A,B,C"   | "B"                 | "2"          |
