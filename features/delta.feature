Feature: Delta
  Client can subscribe and unsubscribe using the incremental variant


  @incremental @non-aggregated @z2
  Scenario Outline: Subscribe to resources one after the other
    Given a target setup with service <service>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <r1> for <service>
    Then the Client receives the resources <r1> and version <v1> for <service>
    When the Client subscribes to resources <r2> for <service>
    Then the Client receives the resources <r2> and version <v1> for <service>
    And the service never responds more than necessary

    Examples:
      | service | resources | r1  | r2  | v1  |
      | "CDS"   | "A,B,C"   | "A" | "B" | "1" |
      | "LDS"   | "D,E,F"   | "D" | "E" | "1" |
