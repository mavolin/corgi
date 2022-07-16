This directory contains integration tests for corgi files.

## How to Run the Tests

Since corgi templates are compiled to Go functions, you must first run the preparation tests,
which will generate the functions then used in the "real" tests.

All preparation tests have the `prepare_integration_test` build tag.

After generating, you may then run the integration test using the `integration_test` build tag.

You may also use the `test.sh` script, which will automatically build corgi, run the tests,
and then will delete the generated functions.

When `test.sh` is run and a test fails, 
it will retain all generated functions to make debugging easier.
To then delete the generated files, run `rm_generated_files.sh`.

## Test Coverage

Because lexer, parser, linker, and writer are not tested individually, 
but all through the integration tests found in here,
in order to obtain the test coverage of corgi,
you must run the integration tests with the `-coverpkg` flag.

You may also run the `codecov.sh` utility script,
which will print the current coverage of the preparation and integration tests to the terminal.

Optionally, you may also specify an `-html` flag to open a browser with a generated HTML report,
for the preparation test.

Lastly, you may also specify the -coverprofile flag to generate to coverage profiles,
`prepare_coverage.txt` and `integration_coverage.txt`, 
that contain coverage information for the preparation and integration tests, respectively.

`-html` and `-coverprofile` are mutually exclusive.
