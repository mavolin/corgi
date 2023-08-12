#!/usr/bin/env bash

find="val, ok := p.parseExpr(rule.expr)"
# shellcheck disable=SC2116 disable=SC2028
replace=$(echo \
  'fmt.Println(strings.Repeat("  ", len(p.vstack))+"+"+rule.name+" ("+p.cur.pos.String()+")")\n' \
  'val, ok := p.parseExpr(rule.expr)\n' \
  'fmt.Println(strings.Repeat("  ", len(p.vstack))+"-"+rule.name)'
)

sed -i "s/$find/$replace/" parser.go