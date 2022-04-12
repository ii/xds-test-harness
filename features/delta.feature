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
