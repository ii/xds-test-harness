Feature: I can give my name and receive a compliment

  Scenario:
    Given a connection to the complimenter
    When I send it a request with my name
    Then I receive a compliment
