Feature: Valid Test Target
  as a test runner
  I want a target I can reach with my program,
  so i can run my tests.

  Rules:
  - target address is 18000

  Scenario:
    Given a target address
    When I attempt to connect to the address
    Then I get a success message
