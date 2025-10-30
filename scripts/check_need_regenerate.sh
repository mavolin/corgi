#!/usr/bin/env bash
set -euo pipefail

files=(
  "escape/charref/chars.go"
  "internal/isstdlib/detector.go"
  "file/switches/functions.go"
)

check_file() {
  local file="$1"
  local package; package="$(dirname "$file")"
  package="$(cd "$package" && pwd)"

  local hash_file; hash_file="$(mktemp)"
  trap 'rm -f "'"$hash_file"'"' EXIT

  sha256sum "$file" > "$hash_file"

  local backup_file; backup_file="$(mktemp)"
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

failed=0
for file in "${files[@]}"; do
  check_file "$file" || failed=1
done
return $failed
