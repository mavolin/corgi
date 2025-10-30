#!/usr/bin/env bash
set -euo pipefail

MODULE="$(go list -m -f "{{.Path}}" all | head -n 1)"

mktmpfile() {
    local tmpfile; tmpfile="$(mktemp)"
    trap 'rm -f "'"$tmpfile"'"' EXIT
    echo "$tmpfile"
}

generated_files() {
  grep -rl '^\w*// Code generated .* DO NOT EDIT\.$'
}

coverage_files=()

coverage_files+=("$(mktmpfile)")
go test -coverprofile "${coverage_files[-1]}" -covermode atomic ./...
gocovmerge "${coverage_files[@]}" > "$1"

grep_tmp="$(mktmpfile)"
for file in $(generated_files); do
  grep -v "^$MODULE/$file:" "$1" > "$grep_tmp"
  mv "$grep_tmp" "$1"
done

