# Contributing

## Code Contributions

### Committing

Please use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for your commits.

##### Types
We use the following types:

- **build**: Changes that affect the build system or external dependencies
- **ci**: changes to our CI configuration files and scripts
- **docs**: changes to the documentation
- **feat**: a new feature
- **fix**: a bug fix
- **perf**: an improvement to performance
- **refactor**: a code change that neither fixes a bug nor adds a feature
- **style**: a change that does not affect the meaning of the code
- **test**: a change to an existing test, or a new test

### Code Style

Make sure all code passes the golangci-lint checks.
If necessary, add a `//nolint:{{name_of_linter}}` directive to the line or block to silence false positives or exceptions.
