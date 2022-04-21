Feature: Delta
  Delta works!

  @wip @incremental @non-aggregated
  Scenario Outline: Subscribe to resources this is cool stuff
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
    Then the Client receives only the resource <r1> and version <v2> for service
    And the service never responds more than necessary
    # And the Client never received resource <r1> at version <v2>

    Examples:
      | service | resources | r1  | r2  | v1  | v2  |
      | "CDS"   | "A,B,C"   | "A" | "B" | "1" | "2" |
      | "LDS"   | "A,B,C"   | "A" | "B" | "1" | "2" |
