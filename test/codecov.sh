function main() {
  ./rm_generated_files.sh

  if [[ $# -eq 0 ]]; then
    go test github.com/mavolin/corgi/test/... \
      -tags prepare_integration_test \
      -coverpkg github.com/mavolin/corgi/... ||
      return $?

    go test github.com/mavolin/corgi/test/... \
      -tags integration_test \
      -coverpkg github.com/mavolin/corgi/pkg/writeutil,github.com/mavolin/corgi/test/... ||
      return $?

    ./rm_generated_files.sh
    return $?
  fi

  if [[ $1 != "-coverprofile" && $1 != "-html" ]]; then
    echo "Usage: ./codecov.sh [-coverprofile | -html]"
    return 1
  fi

  if [[ $1 == "-html" ]]; then
    local coverprofile
    coverprofile="$(mktemp)"

    go test github.com/mavolin/corgi/test/... \
      -tags prepare_integration_test \
      -coverprofile "$coverprofile" \
      -coverpkg github.com/mavolin/corgi/... ||
      return $?

    go tool cover -html "$coverprofile" || return $?

    rm "$coverprofile" || return $?
    ./rm_generated_files.sh || return $?

    return 0
  fi

  go test github.com/mavolin/corgi/test/... \
    -tags prepare_integration_test \
    -coverprofile prepare_coverage.txt \
    -coverpkg github.com/mavolin/corgi/... ||
    return $?

  go test github.com/mavolin/corgi/test/... \
    -tags integration_test \
    -coverprofile integration_coverage.txt \
    -coverpkg github.com/mavolin/corgi/pkg/writeutil,github.com/mavolin/corgi/test/... ||
    return $?

  return 0
}

main "$@"
exit $?
