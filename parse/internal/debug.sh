#!/usr/bin/env bash

find1="func (p \*parser) parseRule(rule \*rule) (any, bool) {"
replace1=$(echo \
  'var nBaseRules int\n' \
  'func (p *parser) parseRule(rule *rule) (any, bool) {' \
)

find2="val, ok := p.parseExprWrap(rule.expr)"
# shellcheck disable=SC2116 disable=SC2028
replace2=$(echo \
  'if n := rule.name; n == "EOC" || n == "EOF" || n == "EOL" || n == "WS" || n == "_" || n == "HORIZ_WS" || n == "VERT_WS" {\n' \
  '    nBaseRules++\n' \
  '} else {\n' \
  '    nBaseRules = 0\n' \
  '}\n' \
  'var val any\n' \
  'var ok bool\n' \
  'if nBaseRules <= 25 {\n' \
  '    fmt.Println(strings.Repeat(" ", len(p.vstack))+"+"+rule.name+" ("+p.cur.pos.String()+")")\n' \
  '    val, ok = p.parseExprWrap(rule.expr)\n' \
  '    fmt.Println(strings.Repeat(" ", len(p.vstack))+"-"+rule.name)\n' \
  '} else {\n' \
  '    val, ok = p.parseExprWrap(rule.expr)\n' \
  '}\n'
)

sed -i "s/$find1/$replace1/" parser.go
sed -i "s/$find2/$replace2/" parser.go