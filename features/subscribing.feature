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
      | xDS   | v1  | resources | expected |
      | "CDS" | "1" | "A,B,C"   | "C,A,B"  |
      | "LDS" | "1" | "D,E,F"   | "F,D,E"  |


  @sotw @non-aggregated @aggregated
  Scenario Outline: [<xDS>] The service should send updates to the client
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client does a wildcard subscription to <xDS>
    Then the Client receives the resources <resources> and version <v1> for <xDS>
    When  the resources <r1> of the <xDS> is updated to version <v2>
    Then the Client receives the resources <resources> and version <v2> for <xDS>
    And the resources <resources> and version <v2> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | v1  | v2  | resources | r1  |
      | "CDS" | "1" | "2" | "A,B,C"   | "A" |
      | "LDS" | "1" | "2" | "D,E,F"   | "D" |


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


  @sotw @non-aggregated @aggregated
  Scenario: [<service>]  When subscribing to specific resources, receive only these resources
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to resources <subset of resources> for <service>
    Then the Client receives the resources <subset of resources> and version <starting version> for <service>
    And the service never responds more than necessary

    Examples:
      | service | starting version | resources | subset of resources |
      | "CDS"   | "1"              | "A,B,C,D" | "B,D"               |
      | "LDS"   | "1"              | "G,B,L,D" | "L,G"               |
      | "RDS"   | "1"              | "A,B"     | "A,B"               |
      | "EDS"   | "1"              | "A,B"     | "A,B"               |


  @CDS @LDS @sotw @non-aggregated @aggregated
  Scenario: [<xDS>] When subscribing to specific resources, receive response when those resources change
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <subset> for <xDS>
    Then the Client receives the resources <subset> and version <v1> for <xDS>
    When the resources <r1> of the <xDS> is updated to version <v2>
    Then the Client receives the resources <subset> and version <v2> for <xDS>
    And the resources <subset> and version <v2> for <xDS> came in a single response
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | subset | r1  | v1  | v2  |
      | "CDS" | "A,B,C,D" | "B,D"  | "B" | "1" | "2" |
      | "LDS" | "G,B,L,D" | "L,G"  | "G" | "1" | "2" |


  @sotw @non-aggregated @aggregated
  Scenario: [<service>] When subscribing to specific resources, receive response when those resources change
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to resources <subset of resources> for <service>
    Then the Client receives the resources <subset of resources> and version <starting version> for <service>
    When the resources <subscribed resource> of the <service> is updated to version <next version>
    Then the Client receives the resources <subscribed resource> and version <next version> for <service>
    And the service never responds more than necessary

    Examples:
      | service | starting version | resources   | subset of resources | subscribed resource | next version |
      | "RDS"   | "1"              | "A,B,C,D"   | "B,D"               | "B"                 | "2"          |
      | "EDS"   | "1"              | "A,B,C,D"   | "B,D"               | "B"                 | "2"          |


  @CDS @LDS @sotw @non-aggregated @aggregated
  Scenario: [<xDS>] When subscribing to resources that don't exist, receive response when they are created
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


  @sotw @non-aggregated @aggregated @wip
  Scenario: [<service>] When subscribing to resources that don't exist, receive response when they are created
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to resources <subset of resources> for <service>
    Then the Client receives the resources <existing subset> and version <starting version> for <service>
    When the resource <chosen resource> is added to the <service> with version <next version>
    Then the Client receives the resources <chosen resource> and version <next version> for <service>
    And the service never responds more than necessary

    Examples:
      | service | starting version | resources   | subset of resources | existing subset | chosen resource | next version |
      | "RDS"   | "1"              | "A,B,C,D"   | "A,Z"               | "A"             | "Z"             | "2"          |
      | "EDS"   | "1"              | "A,B,C,D"   | "A,Z"               | "A"             | "Z"             | "2"          |


  @sotw @aggregated
  Scenario: [<service>] Client can subscribe to multiple services via ADS
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to resources <subset of resources> for <service>
    Then the Client receives the resources <subset of resources> and version <starting version> for <service>
    When the Client subscribes to resources <subset of resources> for <other service>
    Then the Client receives the resources <subset of resources> and version <starting version> for <other service>
    When the resources <subset of resources> of the <service> is updated to version <next version>
    Then the Client receives the resources <subset of resources> and version <next version> for <service>
    # trying out different language for server not responding to acks
    And the service never responds more than necessary

    Examples:
      | service | other service | starting version | resources | subset of resources | next version |
      | "CDS"   | "LDS"         | "1"              | "A,B,C"   | "B"                 | "2"          |
      | "RDS"   | "EDS"         | "1"              | "A,B,C"   | "B"                 | "2"          |
