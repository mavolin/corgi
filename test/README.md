This directory contains integration tests for corgi files.

## How to Run the Tests

Since corgi templates are compiled to Go functions, you must first the preparation tests,
which will generate the functions then used in the real tests.

All preparation tests have the `prepare_integration_test` build tag.
They expect `os.Args[1]` to be the path to the corgi binary.

Then you may run the integration test using the `integration_test` build tag.

You may also use the test.sh script, which will automatically build corgi, run the tests,
and then will delete the generated functions.
