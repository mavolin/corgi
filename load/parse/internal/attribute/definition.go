package attribute

import (
	"regexp"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func Definition() parser.Func[*ast.AttributeDefinition] {
	return func(p *parser.Parser) *ast.AttributeDefinition {
		attr := parser.TryKeywordAt(p, "attr", comment.OrAnyWhitespace())
		if attr == nil {
			return nil
		}

		beforePrefix := p.CloneState()

		var def ast.AttributeDefinition
		def.Attr = attr

		// technically '(' would be a valid attribute name, so check that we
		// don't accidentally consume a '(' as a prefix here
		if !parser.MatchesAnyRune(p, '(') {
			def.Prefix = parser.TryOptional(p, Name(), comment.OrAnyWhitespace())
		}

		def.LParen = parser.TryOptionalRuneAt(p, '(', comment.OrAnyWhitespace())
		if def.LParen == nil {
			if parser.MatchesAnyRune(p, '{') {
				// our prefix is actually a single spec
				def.Prefix = nil
				p.RestoreState(beforePrefix)
			}

			spec := parser.Try(p, Spec())
			if spec != nil {
				def.Specs = []*ast.AttributeSpec{spec}
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message:  "attribute definition: missing attribute name",
					Primary:  quickanno.Expected(p, p.Pos(), "an attribute name"),
					Examples: []diagnostic.Example{{Example: "`attr hx-foo { ... }`"}},
				})
			}
			return &def
		}

		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			s := parser.TryOptional(p, Spec(), nil)
			if s == nil {
				break
			}
			def.Specs = append(def.Specs, s)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}
			parser.Try(p, comment.AndMustEOS())
		}
		def.Specs = slices.Clip(def.Specs)

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "attribute definition: unexpected runes before closing parenthesis"
			p.CaptureError(err)
		}

		def.RParen = parser.TryOptionalRuneAt(p, ')', nil)
		if def.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute definition: unclosed parenthesis",
				Primary: quickanno.Expected(p, *def.LParen, "expected a `)` for the opening `(` here"),
			})
		}

		return &def
	}
}

func Spec() parser.Func[*ast.AttributeSpec] {
	return func(p *parser.Parser) *ast.AttributeSpec {
		selector := parser.Try(p, Selector())
		if selector == nil {
			return nil
		}

		var spec ast.AttributeSpec
		spec.Selector = selector

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		spec.Ruleset = parser.Try(p, Ruleset())
		if spec.Ruleset == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "attribute spec: missing ruleset",
				Primary:  quickanno.Expected(p, p.Pos(), "an attribute ruleset"),
				Examples: []diagnostic.Example{{Example: "`{ * innocuous }`"}},
			})
		}
		return &spec
	}
}

func Ruleset() parser.Func[*ast.AttributeRuleset] {
	return func(p *parser.Parser) *ast.AttributeRuleset {
		lBrace := parser.TryRuneAt(p, '{')
		if lBrace == nil {
			return nil
		}

		var rs ast.AttributeRuleset
		rs.LBrace = lBrace

		rs.List = make([]*ast.AttributeRule, 0, 64)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			r := parser.TryOptional(p, Rule(), nil)
			if r == nil {
				break
			}
			parser.Try(p, comment.AndForceEOS())
			rs.List = append(rs.List, r)
		}
		rs.List = slices.Clip(rs.List)

		rs.RBrace = parser.TryRuneAt(p, '}')
		if rs.RBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute ruleset: unclosed brace",
				Primary: quickanno.Expected(p, *rs.LBrace, "expected a `}` for the opening `{` here"),
			})
		}

		return &rs
	}
}

func Rule() parser.Func[*ast.AttributeRule] {
	return func(p *parser.Parser) *ast.AttributeRule {
		selector := parser.Try(p, ElementSelector())
		if selector == nil {
			return nil
		}

		var r ast.AttributeRule
		r.Selector = selector
		parser.TrySkip(p, comment.OrAnyWhitespace())

		r.Type = parser.Try(p, TypeName())
		if r.Type == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "attribute rule: missing type",
				Primary:  quickanno.Expected(p, p.Pos(), "a type name"),
				Examples: []diagnostic.Example{{Example: "`div innocuous`"}},
			})
		}
		return &r
	}
}

func Selector() parser.Func[ast.AttributeSelector] {
	return func(p *parser.Parser) ast.AttributeSelector {
		if bs := parser.Try(p, BasicSelector()); bs != nil {
			return bs
		} else if rs := parser.Try(p, RegexpSelector()); rs != nil {
			return rs
		}
		return nil
	}
}

func BasicSelector() parser.Func[*ast.BasicAttributeSelector] {
	return func(p *parser.Parser) *ast.BasicAttributeSelector {
		pos := p.Pos()
		name := parser.Try(p, Name())
		if name == nil {
			return nil
		}

		var s ast.BasicAttributeSelector
		s.Position = &pos
		s.Name = name.Name

		s.Wildcard = strings.HasSuffix(s.Name, "*")
		if s.Wildcard {
			s.Name = s.Name[:len(s.Name)-1]
			if s.Name == "" {
				p.CaptureError(&diagnostic.Diagnostic{
					Message:     "basic attribute selector: only wildcard",
					Primary:     quickanno.Expected(p, *s.Position, "an attribute name before the wildcard"),
					Explanation: "A basic attribute selector mustn't consist solely of a wildcard.",
				})
			}
		}

		return &s
	}
}

func RegexpSelector() parser.Func[*ast.RegexpAttributeSelector] {
	return func(p *parser.Parser) *ast.RegexpAttributeSelector {
		re := parser.TryTokenAt(p, "'regexp")
		if re == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		lParen := parser.TryRuneAt(p, '(')
		if lParen == nil {
			return nil
		}
		parser.TrySkip(p, whitespace.Any())

		var s ast.RegexpAttributeSelector
		s.Regexp = re
		s.LParen = lParen

		s.Raw = parser.Try(p, golang.StringLit())
		if s.Raw == nil || s.Raw.Contents == "" {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "regexp attribute selector: missing regexp",
				Primary:  quickanno.Expected(p, p.Pos(), "a string containing a regular expression"),
				Examples: []diagnostic.Example{{Example: "'regexp(`hx-\\d{3}`)"}},
			})
		} else {
			var err error
			s.Compiled, err = regexp.Compile(s.Raw.Unquote())
			if err != nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "invalid regular expression: " + err.Error(),
					Primary: quickanno.Expected(p, s.Raw.Start(), "a valid regular expression"),
				})
			}
		}

		parser.TrySkip(p, whitespace.Any())
		s.RParen = parser.TryRuneAt(p, ')')
		if s.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "regexp attribute selector: missing closing parenthesis",
				Primary: quickanno.Expected(p, *s.LParen, "a closing parenthesis"),
			})
		}

		return &s
	}
}

func ElementSelector() parser.Func[ast.ElementSelector] {
	return func(p *parser.Parser) ast.ElementSelector {
		if ws := parser.Try(p, WildcardElementSelector()); ws != nil {
			return ws
		} else if ls := parser.Try(p, ListElementSelector()); ls != nil {
			return ls
		}
		return nil
	}
}

func WildcardElementSelector() parser.Func[*ast.WildcardElementSelector] {
	return func(p *parser.Parser) *ast.WildcardElementSelector {
		asterisk := parser.TryRuneAt(p, '*')
		if asterisk == nil {
			return nil
		}

		return &ast.WildcardElementSelector{Asterisk: asterisk}
	}
}

func ListElementSelector() parser.Func[*ast.ListElementSelector] {
	return func(p *parser.Parser) *ast.ListElementSelector {
		list := parser.Try(p, list.CommaList("element name", "element names", elementReference))
		if len(list) == 0 {
			return nil
		}

		return &ast.ListElementSelector{List: list}
	}
}

var elementReference parser.Func[*ast.ElementReference]

func SetElementReference(f parser.Func[*ast.ElementReference]) {
	elementReference = f
}
