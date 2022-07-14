function main() {
    if [[ $# -eq 0 ]]; then
      go test github.com/mavolin/corgi/test/... \
        -tags prepare_integration_test \
        -coverpkg=github.com/mavolin/corgi/corgi/...,github.com/mavolin/corgi/writer ||
      return $?

      ./rm_generated_files.sh
      return $?
    fi

    if [[ $1 != "-coverprofile" && $1 != "-html" ]];then
      echo "Usage: ./codecov.sh [-coverprofile [file]|-html]"
      return 1
    fi

    local coverprofile="cover.out"
    if [[ $# -eq 2 ]]; then
      coverprofile=$2
    fi

    go test github.com/mavolin/corgi/test/... \
      -tags prepare_integration_test \
      -coverprofile=${coverprofile} \
      -coverpkg=github.com/mavolin/corgi/corgi/...,github.com/mavolin/corgi/writer ||
      return $?

    if [[ $1 == "-html" ]]; then
      go tool cover -html=cover.out || return $?
      rm cover.out || return $?
      ./rm_generated_files.sh || return $?
    fi

    return 0
}

main "$@"
exit $?
