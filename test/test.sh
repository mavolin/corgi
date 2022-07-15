#!/usr/bin/env bash

function main() {
  ./rm_generated_files.sh
  test
  local test_status=$?

  # don't delete on test failure to allow debugging
  if [ $test_status -eq 0 ]; then
    ./rm_generated_files.sh
  fi

  exit $test_status
}

function test() {
  go test github.com/mavolin/corgi/test/... -count=1 -tags prepare_integration_test || return $?

  go test github.com/mavolin/corgi/test/... -count=1 -tags integration_test  || return $?
  return $?
}

main
