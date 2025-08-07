#!/usr/bin/env bash
set -euo pipefail

main() {
  local failed=0
  check_file "escape/charref/chars.go" || failed=1
  check_file "internal/isstdlib/detector.go" || failed=1
  return $failed
}

check_file() {
  local file="$1"
  local package; package="$(dirname "$file")" || return $?
  package="$(cd "$package" && pwd)" || return $?

  local hash_file; hash_file="$(mktemp)" || return $?
  trap 'rm -f "'"$hash_file"'"' EXIT

  sha256sum "$file" > "$hash_file"

  local backup_file; backup_file="$(mktemp)" || return $?
  trap 'rm -f "'"$backup_file"'"' EXIT

  cp "$file" "$backup_file"

  local go_generate_output
  if ! go_generate_output="$(go generate "$package" 2>&1)"; then
    echo "$go_generate_output"
    return 1
  fi

  sha256sum --strict --check "$hash_file" || {
    cp "$backup_file" "$file"
    return 1
  }

  return 0
}

main
