@wip
Feature: Fetching Resources with LDS and CDS
  Client can do wildcard subscriptions or normal subscriptions
  and receive updates when any subscribed resources change.

  These features come from this list of test cases:
  https://docs.google.com/document/d/19oUEt9jSSgwNnvZjZgaFYBHZZsw52f2MwSo6LWKzg-E

  Scenario Outline: The service should send all resources on a wildcard request.
    Given a target setup with <service>, <resources>, and <starting version>
    When the Client does a wildcard subscription to <service>
    Then the Client receives the <expected resources> and <starting version> for <service>
    And the Client sends an ACK to which the <service> does not respond

    Examples:
      # Steps 3 and 5 should fail
      | service | starting version | resources | expected resources |
      | "CDS"   | "1"              | "A,B,C"   | "A,B,C"            |
      | "CDS"   | "1"              | "A,B,C"   | "C,A,B"            |
      | "CDS"   | "1"              | "A,B,C"   | "B,A,D"            |
      | "LDS"   | "1"              | "D,E,F"   | "D,E,F"            |
      | "LDS"   | "1"              | "D,E,F"   | "F,A,I,L"          |
      | "LDS"   | "1"              | "D,E,F"   | "F,D,E"            |
