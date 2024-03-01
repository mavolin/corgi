#!/usr/bin/env bash

# The single corgi.peg is getting long.
# Since pigeon itself does not offer to read grammars from multiple files, I
# wrote this shell script which combines all .peg files into a single full.peg
# file, which is then compiled using pigeon.
# Args passed to this script a redirected to the pigeon command, however, the
# input file must not be specified.

MAIN_FILE=main.peg # The file containing the initializer and first rule.
OUT_FILE=full.peg # The name of the output file, to which all other .peg files appended. Deleted after successful compilation.

cat "$MAIN_FILE" > "$OUT_FILE" || exit $?

for f in *.peg; do
  if [[ "$f" != "$MAIN_FILE" ]] && [[ "$f" != "$OUT_FILE" ]]; then
    cat "$f" >> "$OUT_FILE" || exit $?
    echo >> "$OUT_FILE" || exit $? # newline
  fi
done

# shellcheck disable=SC2068
pigeon $@ full.peg || exit $?

rm "$OUT_FILE"
