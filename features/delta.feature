Feature: Delta
  Client can subscribe and unsubscribe using the incremental variant


  @incremental @non-aggregated @aggregated
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
      | "RDS"   | "D,E,F"   | "D" | "E" | "1" |
      | "EDS"   | "D,E,F"   | "D" | "E" | "1" |


 @incremental @non-aggregated @aggregated
  Scenario Outline: When a resource is updated, receive response for only that resource
    Given a target setup with service <service>, resources <resources>, and starting version <v1>
     When the Client subscribes to resources <resources> for <service>
     Then the Client receives the resources <resources> and version <v1> for <service>
     When the resource <r1> of service <service> is updated to version <v2>
     Then the Client receives the resources <r1> and version <v2> for <service>
      And for service <service>, no resource other than <r1> has same version or nonce
      And the service never responds more than necessary

    Examples:
      | service | resources | r1  | v1  | v2  |
      | "LDS"   | "A,B,C"   | "A" | "1" | "2" |
      | "CDS"   | "A,B,C"   | "A" | "1" | "2" |
      | "RDS"   | "A,B,C"   | "A" | "1" | "2" |
      | "EDS"   | "A,B,C"   | "A" | "1" | "2" |

 @incremental @non-aggregated @aggregated
 Scenario: Client is told if resource does not exist, and is notified if it is created
   Given a target setup with service <service>, resources <r1>, and starting version <v1>
    When the Client subscribes to resources <resources> for <service>
    Then the Client receives the resources <r1> and version <v1> for <service>
     And for service <service>, no resource other than <r1> has same version or nonce
    # And the Delta Client is told <r2> does not exist
    When the resource <r2> is added to the <service> with version <v1>
    Then the Client receives the resources <r1> and version <v1> for <service>
     And for service <service>, no resource other than <r2> has same nonce
     And the service never responds more than necessary

   Examples:
     | service | v1  | resources | r1  | r2  |
     | "CDS"   | "1" | "A,B"     | "A" | "B" |
     | "LDS"   | "1" | "D,E"     | "D" | "E" |
     | "RDS"   | "1" | "D,E"     | "D" | "E" |
     | "EDS"   | "1" | "D,E"     | "D" | "E" |


 @incremental @non-aggregated @zz
 Scenario: Client is told when a resource is removed via removed_resources field
   Given a target setup with service <service>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <resources> for <service>
    Then the Client receives the resources <resources> and version <v1> for <service>
    When the resource <r1> is removed from the <service>
    Then the Client receives notice that resource <r1> was removed for service <service>
    When the resource <r1> is added to the <service> with version <v2>
    Then the Client receives the resources <r1> and version <v2> for <service>
     And for service <service>, no resource other than <r1> has same version or nonce
     And the service never responds more than necessary

    Examples:
      | service | resources | r1  | v1  | v2  |
      | "CDS"   | "A,B"     | "B" | "1" | "2" |
      | "LDS"   | "D,E"     | "E" | "1" | "2" |
      | "RDS"   | "D,E"     | "E" | "1" | "2" |
      | "EDS"   | "D,E"     | "E" | "1" | "2" |
