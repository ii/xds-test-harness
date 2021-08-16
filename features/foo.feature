Feature: Foo the Things
  A dummy feature file to double check our test suite is reusing steps accordingly

  Scenario: A wildcard foo request should be awesome
    Given a target setup with the following "foosources":
    ```
    foobles:
    - name: foo
      version: 1
    ```
    When the Runner says "foo"
    Then the Runner receives the following "foobles":
    ```
    - name: foo
      version: 1
    ```
