Feature: Delta
  Delta works!

  @incremental @non-aggregated
  Scenario Outline: Client can subscribe to resources using incremental variant
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to resources <resources> for <service>
    Then the Client receives the resources <resources> and version <starting version> for <service>
    And the service never responds more than necessary

    Examples:
      | service | starting version | resources |
      | "CDS"   | "1"              | "A,B,C"   |
      | "LDS"   | "1"              | "D,E,F"   |


  @incremental @non-aggregated
  Scenario Outline: Subscribe to resources one after the other
    Given a target setup with service <service>, resources <resources>, and starting version <starting version>
    When the Client subscribes to resources <r1> for <service>
    Then the Client receives the resources <r1> and version <starting version> for <service>
    When the Client subscribes to resources <r2> for <service>
    Then the Client receives the resources <r2> and version <starting version> for <service>
    And the service never responds more than necessary

    Examples:
      | service | starting version | resources | r1  | r2  |
      | "CDS"   | "1"              | "A,B,C"   | "A" | "B" |
      | "LDS"   | "1"              | "D,E,F"   | "D" | "E" |


  @incremental @non-aggregated
  Scenario Outline: When a resource is updated, receive response for only that resource
    Given a target setup with service <service>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <resources> for <service>
    Then the Client receives the resources <resources> and version <v1> for <service>
    When the resource <r1> of service <service> is updated to version <v2>
    Then the Client receives only the resource <r1> and version <v2> for service <service>
    And the service never responds more than necessary

    Examples:
      | service | resources | r1  | v1  | v2  |
      | "CDS"   | "A,B,C"   | "A" | "1" | "2" |
      | "LDS"   | "A,B,C"   | "A" | "1" | "2" |


  @incremental @non-aggregated
  Scenario: Client is told if resource does not exist, and is notified if it is created
    Given a target setup with service <service>, resources <r1>, and starting version <v1>
    When the Client subscribes to resources <resources> for <service>
    Then the Delta Client receives only the resource <r1> and version <v1> for service <service>
    # And the Delta Client is told <r2> does not exist
    When the resource <r2> is added to the <service> with version <v1>
    Then the Delta Client receives only the resource <r2> and version <v1> for service <service>
    And the service never responds more than necessary

    Examples:
      | service | v1  | resources | r1  | r2  |
      | "CDS"   | "1" | "A,B"     | "A" | "B" |
      | "LDS"   | "1" | "D,E"     | "D" | "E" |


  @incremental @non-aggregated
  Scenario: Client is told when a resource is removed via removed_resources field
    Given a target setup with service <service>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <resources> for <service>
    Then the Client receives the resources <resources> and version <v1> for <service>
    When the resource <r1> is removed from the <service>
    Then the Delta Client receives notice that resource <r1> was removed for service <service>
    When the resource <r1> is added to the <service> at version <v2>
    Then the Delta Client receives only the resource <r1> and version <v2> for service <service>
    And the service never responds more than necessary

    Examples:
      | service | resources | r1  | v1  | v2  |
      | "CDS"   | "A,B"     | "B" | "1" | "2" |
      | "LDS"   | "D,E"     | "E" | "1" | "2" |

  @incremental @non-aggregated @z2
  Scenario: Client can incrementally unsubscribe from resources
    Given a target setup with service <service>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <resources> for <service>
    Then the Client receives the resources <resources> and version <v1> for <service>
    When the Client unsubscribes from resource <r1> for service <service>
    And the resource <r1> of service <service> is updated to version <v2>
    Then the Delta client does not receive resource <r1> of service <service> at version <v2>
    When the Client unsubscribes from resource <r2> for service <service>
    And the resource <r2> of service <service> is updated to version <v2>
    Then the Delta client does not receive resource <r2> of service <service> at version <v2>
    And the service never responds more than necessary

    Examples:
      | service | resources | r1  | r2  | v1  | v2  |
      | "CDS"   | "A,B"     | "B" | "A" | "1" | "2" |
      | "LDS"   | "D,E"     | "E" | "D" | "1" | "2" |
