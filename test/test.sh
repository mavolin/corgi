#!/usr/bin/env bash

function main() {
  test
  local test_status=$?

  # don't delete on test failure to allow debugging
  if [ $test_status -eq 0 ]; then
    ./remove_generated_files.sh
  fi

  exit $test_status
}

function test() {
  go test github.com/mavolin/corgi/test/... -tags prepare_integration_test || return $?

  go test github.com/mavolin/corgi/test/... -tags integration_test  || return $?
  return $?
}

main
