Feature: Unsubscribing to Resources
  After subscribing to a set of resources, a client can unbuscribe from any
  or all of these resources. The client should no longer receive updates from
  resources they've unsubcribed to.

  These features come from this list of test cases:
  https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E

  @sotw @non-aggregated @aggregated @zz
  Scenario: [<xDS>] Client can unsubcribe from some resources
    # This test does not check if the final results are only the subscribed resources
    # it is valid(though not desired) for a server to send more than is requested.
    # So the test will pass if client subscribes to A,B,C, unsubscribes from B,C and gets A,B,C back.
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <subset> and version <v1> for <xDS>
    When the Client updates subscription to a resource(<r1>) of <xDS> with version <v1>
    And the resource <r1> of service <xDS> is updated to version <v2>
    Then the Client receives the resources <r1> and version <v2> for <xDS>
    And the resources <r1> and version <v2> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | resources   | subset  | r1    | v1  | v2  |
      | "CDS" | "F,A,B,C,D" | "C,A,B" | "A" | "1" | "2" |
      | "LDS" | "G,B,L,D"   | "B,D,L" | "B" | "1" | "2" |


  @sotw @non-aggregated @aggregated
  Scenario: [<xDS>] Client can unsubcribe from some resources
    # difference from test above is use of the word ONLY in the final THEN step
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <subset> and version <v1> for <xDS>
    When the Client updates subscription to a resource(<r1>) of <xDS> with version <v1>
    And the resource <r1> of service <xDS> is updated to version <v2>
    Then the Client receives only the resource <r1> and version <v2> for the service <xDS>
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | subset | r1  | v1  | v2  |
      | "RDS" | "A,B,C,D" | "B,D"  | "B" | "1" | "2" |
      | "EDS" | "A,B,C,D" | "B,D"  | "B" | "1" | "2" |


  @sotw @non-aggregated @aggregated
  Scenario: [<xDS>] Client can unsubscribe from all resources
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <subset> and version <v1> for <xDS>
    When the Client unsubscribes from all resources for <xDS>
    And the resource <r1> of service <xDS> is updated to version <v2>
    Then the Client does not receive any message from <xDS>

    Examples:
      | xDS   | resources | subset | r1  | v1  | v2  |
      | "CDS" | "A,B,C,D" | "A,B"  | "B" | "1" | "2" |
      | "RDS" | "A,B,C,D" | "A,B"  | "B" | "1" | "2" |
      | "LDS" | "A,B,C,D" | "A,B"  | "B" | "1" | "2" |
      | "EDS" | "A,B,C,D" | "A,B"  | "B" | "1" | "2" |
