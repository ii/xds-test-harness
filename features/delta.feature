Feature: Delta
  Client can subscribe and unsubscribe using the incremental variant


  @incremental @non-aggregated @aggregated
  Scenario Outline: [<xDS>] Subscribe to resources one after the other
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <r1> for <xDS>
    Then the Client receives the resources <r1> and version <v1> for <xDS>
    When the Client subscribes to resources <r2> for <xDS>
    Then the Client receives the resources <r2> and version <v1> for <xDS>
    And the service never responds more than necessary

    Examples:
      | xDS   | resources | r1  | r2  | v1  |
      | "CDS" | "A,B,C"   | "A" | "B" | "1" |
      | "LDS" | "D,E,F"   | "D" | "E" | "1" |
      | "RDS" | "D,E,F"   | "D" | "E" | "1" |
      | "EDS" | "D,E,F"   | "D" | "E" | "1" |


 @incremental @non-aggregated @aggregated
 Scenario Outline: [<xDS>] When a resource is updated, receive response for only that resource
    Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
     When the Client subscribes to resources <resources> for <xDS>
     Then the Client receives the resources <resources> and version <v1> for <xDS>
     When the resource <r1> of service <xDS> is updated to version <v2>
     Then the Client receives the resources <r1> and version <v2> for <xDS>
      And for service <xDS>, no resource other than <r1> has same version or nonce
      And the service never responds more than necessary

    Examples:
      | xDS   | resources | r1  | v1  | v2  |
      | "LDS" | "A,B,C"   | "A" | "1" | "2" |
      | "CDS" | "A,B,C"   | "A" | "1" | "2" |
      | "RDS" | "A,B,C"   | "A" | "1" | "2" |
      | "EDS" | "A,B,C"   | "A" | "1" | "2" |

 @incremental @non-aggregated @aggregated
 Scenario Outline: [<xDS>] Client is told if resource does not exist, and is notified if it is created
   Given a target setup with service <xDS>, resources <r1>, and starting version <v1>
    When the Client subscribes to resources <resources> for <xDS>
    Then the Client receives the resources <r1> and version <v1> for <xDS>
     And for service <xDS>, no resource other than <r1> has same version or nonce
    # And the Delta Client is told <r2> does not exist
    When the resource <r2> is added to the <xDS> with version <v1>
    Then the Client receives the resources <r2> and version <v1> for <xDS>
     And for service <xDS>, no resource other than <r2> has same nonce
     And the service never responds more than necessary

   Examples:
     | xDS   | v1  | resources | r1  | r2  |
     | "CDS" | "1" | "A,B"     | "A" | "B" |
     | "LDS" | "1" | "D,E"     | "D" | "E" |
     | "RDS" | "1" | "D,E"     | "D" | "E" |
     | "EDS" | "1" | "D,E"     | "D" | "E" |


 @incremental @non-aggregated @aggregated
 Scenario Outline: [<xDS>] Client is told when a resource is removed via removed_resources field
   Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <resources> for <xDS>
    Then the Client receives the resources <resources> and version <v1> for <xDS>
    When the resource <r1> is removed from the <xDS>
    Then the Client receives notice that resource <r1> was removed for service <xDS>
    When the resource <r1> is added to the <xDS> with version <v2>
    Then the Client receives the resources <r1> and version <v2> for <xDS>
     And for service <xDS>, no resource other than <r1> has same version or nonce
     And the service never responds more than necessary

   Examples:
     | xDS   | resources | r1  | v1  | v2  |
     | "CDS" | "A,B"     | "B" | "1" | "2" |
     | "LDS" | "D,E"     | "E" | "1" | "2" |
     | "RDS" | "D,E"     | "E" | "1" | "2" |
     | "EDS" | "D,E"     | "E" | "1" | "2" |


 @incremental @non-aggregated @aggregated
 Scenario Outline: [<xDS>] Client can incrementally unsubscribe from resources
   Given a target setup with service <xDS>, resources <resources>, and starting version <v1>
    When the Client subscribes to resources <resources> for <xDS>
    Then the Client receives the resources <resources> and version <v1> for <xDS>
    When the Client unsubscribes from resource <r1> for service <xDS>
     And the resource <r1> of service <xDS> is updated to version <v2>
    Then the client does not receive resource <r1> of service <xDS> at version <v2>
    When the Client unsubscribes from resource <r2> for service <xDS>
     And the resource <r2> of service <xDS> is updated to version <v2>
    Then the client does not receive resource <r2> of service <xDS> at version <v2>
     And the service never responds more than necessary

   Examples:
     | xDS   | resources | r1  | r2  | v1  | v2  |
     | "CDS" | "A,B"     | "B" | "A" | "1" | "2" |
     | "LDS" | "D,E"     | "E" | "D" | "1" | "2" |
     | "RDS" | "D,E"     | "E" | "D" | "1" | "2" |
     | "EDS" | "D,E"     | "E" | "D" | "1" | "2" |

  @incremental @aggregated
  Scenario Outline: [<services>] Client can subscribe to multiple services via ADS
    Given a target setup with multiple services <services>, each with resources <resources>, and starting version <v1>
     When the Client subscribes to resources <r1> for <xDS>
      And the Client subscribes to resources <r1> for <xDS2>
     Then the Client receives the resources <r1> and version <v1> for <xDS>
      And the Client receives the resources <r1> and version <v1> for <xDS2>
     When the resource <r1> of service <xDS> is updated to version <v2>
     Then the Client receives the resources <r1> and version <v2> for <xDS>
     When the resource <r1> of service <xDS2> is updated to version <v2>
     Then the Client receives the resources <r1> and version <v2> for <xDS2>
      And the service never responds more than necessary

    Examples:
      | services  | xDS   | xDS2  | resources | r1  | v1  | v2  |
      | "CDS,LDS" | "CDS" | "LDS" | "A,B,C"   | "B" | "1" | "2" |
      | "LDS,CDS" | "LDS" | "CDS" | "A,B,C"   | "B" | "1" | "2" |
      | "RDS,EDS" | "RDS" | "EDS" | "A,B,C"   | "B" | "1" | "2" |
      | "EDS,RDS" | "EDS" | "RDS" | "A,B,C"   | "B" | "1" | "2" |
