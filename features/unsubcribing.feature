Feature: Unsubscribing to Resources
  After subscribing to a set of resources, a client can unbuscribe from any
  or all of these resources. The client should no longer receive updates from
  resources they've unsubcribed to.

  These features come from this list of test cases:
  https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E

  @CDS @LDS @sotw @non-aggregated @aggregated
  Scenario: [<service>] Client can unsubcribe from some resources
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


  @sotw @non-aggregated @aggregated
  Scenario: [<service>] Client can unsubcribe from some resources
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


  @sotw @non-aggregated @aggregated @skip
  Scenario: [<service>] Client can unsubscribe from all resources
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
    Scenario: [<service>] Client can subscribe to multiple services via ADS
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
