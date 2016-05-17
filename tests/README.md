VIC Integration & Functional Test Suite
=======

To run the deprecated tests:

1. Integration tests can be run by calling `make integration-tests` from the project's root directory.

To run these tests:  

1. Build the docker image locally from the tests directory:  

 * `docker build -t vic-integration-test --no-cache - < Dockerfile.vic-integration-test`

2. Execute drone from the projects root directory:  

 * `drone exec --yaml ".drone-e2e.yml"`

Find the documentation for each of the tests here:
-
###[Test Suite Documentation](test-cases/TestGroups.md)
