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
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to a subset of resources,<subset of resources>, for <service>
    Then the Client receives the resources <subset of resources> and version <starting version> for <service>
    When the Client updates subscription to a resource(<resource from subset>) of <service> with version <starting version>
    And  the resources <resource from subset> of the <service> is updated to version <next version>
    Then the Client receives the resources <resource from subset> and version <next version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | resource from subset |   next version  |
      | "CDS"   | "1"              | "F,A,B,C,D" | "C,A,B"             | "A"                |   "2"           |
      | "LDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                  |   "2"           |


  @sotw @non-aggregated @aggregated @active
  Scenario: [<service>] Client can unsubscribe from some resources
    # difference from test above is use of the word ONLY in the final THEN step
    # This currently does not pass for go-control-plane
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to a subset of resources,<subset of resources>, for <service>
    Then the Client receives the resources <subset of resources> and version <starting version> for <service>
    When the Client updates subscription to a resource(<resource from subset>) of <service> with version <starting version>
    And  the resources <resource from subset> of the <service> is updated to version <next version>
    Then the Client receives only the resource <resource from subset> and version <next version>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      | service | starting version | resources   | subset of resources | resource from subset |   next version  |
      | "RDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                  |   "2"           |
      | "EDS"   | "1"              | "G,B,L,D"   | "B,D"               | "B"                  |   "2"           |


  @sotw @non-aggregated @aggregated
  Scenario: [<service>] Client can unsubscribe from all resources
    # This is not working currently, the unsusbcribe is not registered,
    # neither as an unsubscribe nor a new wildcard request
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to a subset of resources,<subset of resources>, for <service>
    Then the Client receives the resources <subset of resources> and version <starting version> for <service>
    When the Client unsubscribes from all resources for <service>
    And the resources <subset of resources> of the <service> is updated to version <next version>
    Then the Client does not receive any message from <service>

    Examples:
      | service | starting version | resources   | subset of resources | next version |
      | "CDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |
      | "LDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |
      | "RDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |
      | "EDS"   | "1"              | "A,B,C,D"   | "A,B"               | "2"          |


