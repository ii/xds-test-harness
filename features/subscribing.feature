Feature: Subscribing to Resources
  Client can do wildcard subscriptions or normal subscriptions
  and receive updates when any subscribed resources change.

  These features come from this list of test cases:
  https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E

  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] The service should send all resources on a wildcard request.
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client does a wildcard subscription to <xDS>
    Then the Client receives the resources <expected> and version <v1> for <xDS>
    And the resources <expected> and version <v1> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | expected | v1  |
      | "CDS" | "A,B,C"   | "C,A,B"  | "1" |
      | "LDS" | "D,E,F"   | "F,D,E"  | "1" |


  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] The service should send updates to the client
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client does a wildcard subscription to <xDS>
    Then the Client receives the resources <resources> and version <v1> for <xDS>
    When the resource <r1> of service <xDS> is updated to version <v2>
    Then the Client receives the resources <resources> and version <v2> for <xDS>
    And the resources <resources> and version <v2> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | r1  | v1  | v2  |
      | "CDS" | "A,B,C"   | "A" | "1" | "2" |
      | "LDS" | "D,E,F"   | "D" | "1" | "2" |


  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] Wildcard subscriptions receive updates when new resources are added
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client does a wildcard subscription to <xDS>
    Then the Client receives the resources <resources> and version <v1> for <xDS>
    When the resource <r1> is added to the <xDS> with version <v2>
    Then the Client receives the resources <expected> and version <v2> for <xDS>
    And the resources <expected> and version <v2> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | r1  | expected  | v1  | v2  |
      | "CDS" | "A,B,C"   | "D" | "A,B,C,D" | "1" | "2" |
      | "LDS" | "D,E,F"   | "G" | "D,E,F,G" | "1" | "2" |


  @sotw @incremental @non-aggregated @aggregated
  Scenario Outline: [<xDS>]  When subscribing to specific resources, receive only these resources
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <subset> and version <v1> for <xDS>
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | subset | v1  |
      # | "CDS" | "A,B,C,D" | "B,D"  | "1" |
      | "LDS" | "G,B,L,D" | "L,G"  | "1" |
      | "RDS" | "B,A"   | "B,A"  | "1" |
      # | "EDS" | "A,B"     | "A,B"  | "1" |


  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] When subscribing to specific resources, receive response when those resources change
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <subset> and version <v1> for <xDS>
    When the resource <r1> of service <xDS> is updated to version <v2>
    Then the Client receives the resources <subset> and version <v2> for <xDS>
    And the resources <subset> and version <v2> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | subset | r1  | v1  | v2  |
      | "CDS" | "A,B,C,D" | "B,D"  | "B" | "1" | "2" |
      | "LDS" | "G,B,L,D" | "L,G"  | "G" | "1" | "2" |


  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] When subscribing to specific resources, receive response when those resources change
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <subset> and version <v1> for <xDS>
    When the resource <r1> of service <xDS> is updated to version <v2>
    Then the Client receives the resources <r1> and version <v2> for <xDS>
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | subset | r1  | v1  | v2  |
      | "RDS" | "A,B,C,D" | "B,D"  | "B" | "1" | "2" |
      | "EDS" | "A,B,C,D" | "B,D"  | "B" | "1" | "2" |


  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] When subscribing to resources that don't exist, receive response when they are created
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <existing subset> and version <v1> for <xDS>
    When the resource <r1> is added to the <xDS> with version <v2>
    Then the Client receives the resources <subset> and version <v2> for <xDS>
    And the resources <subset> and version <v2> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | subset  | existing subset | r1  | v1  | v2  |
      | "CDS" | "A,B,C"   | "A,Z"   | "A"             | "Z" | "1" | "2" |
      | "LDS" | "G,B,D"   | "B,D,X" | "B,D"           | "X" | "1" | "2" |


  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] When subscribing to resources that don't exist, receive response when they are created
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <existing subset> and version <v1> for <xDS>
    When the resource <r1> is added to the <xDS> with version <v2>
    Then the Client receives the resources <r1> and version <v2> for <xDS>
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | subset | existing subset | r1  | v1  | v2  |
      | "RDS" | "A,B,C,D" | "A,Z"  | "A"             | "Z" | "1" | "2" |
      | "EDS" | "A,B,C,D" | "A,Z"  | "A"             | "Z" | "1" | "2" |


  @sotw @aggregated
  Scenario Outline: [<xDS>] Client can subscribe to multiple services via ADS
    Given a target setup with multiple services <services>, each with resources <resources>, and starting version <v1>
    When the Client subscribes to resources <r1> for <xDS>
    Then the Client receives the resources <r1> and version <v1> for <xDS>
    When the Client subscribes to resources <r1> for <xds2>
    Then the Client receives the resources <r1> and version <v1> for <xds2>
    When the resource <r1> of service <xDS> is updated to version <v2>
    Then the Client receives the resources <r1> and version <v2> for <xDS>
    And the service never responds more than necessary

    Examples:
      | services  | xDS   | xds2  | resources | r1  | v1  | v2  |
      | "CDS,LDS" | "CDS" | "LDS" | "A,B,C"   | "B" | "1" | "2" |
      | "RDS,EDS" | "RDS" | "EDS" | "A,B,C"   | "B" | "1" | "2" |
