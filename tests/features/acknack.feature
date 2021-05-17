Feature: Conformance ACK/NACK
  Discovery Request's and Responses should follow the behaviour outlined in the
  API docs.

  Background:
    Given an Adapter located at "localhost:6767"
    And a Target located at "localhost:18000"
    And a Shim located at "localhost:17000"

  Scenario:
    Given a Target with clusters specified with yaml:
     ```
     name: test_config
     spec:
       clusters:
       - name: foo
     ```
     When I send a wildcard request to the CDS
     Then I get a response containing "foo".
