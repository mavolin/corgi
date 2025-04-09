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
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func Definition() parser.Func[*ast.AttributeDefinition] {
	return func(p *parser.Parser) (*ast.AttributeDefinition, *diagnostic.Diagnostic) {
		var def ast.AttributeDefinition

		def.Attr = parser.TryKeywordAt(p, "attr", comment.OrAnyWhitespace())
		if def.Attr == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute definition",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute definition"),
			}
		}

		beforePrefix := p.CloneState()

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

			s := parser.Must(p, Spec())
			if s != nil {
				def.Specs = []*ast.AttributeSpec{s}
			}
			return &def, nil
		}

		def.Specs = make([]*ast.AttributeSpec, 0, 64)
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
			parser.MustSkip(p, comment.AndEOS())
		}
		if len(def.Specs) == 0 {
			def.Specs = nil
		} else {
			def.Specs = slices.Clip(def.Specs)
		}

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

		return &def, nil
	}
}

func Spec() parser.Func[*ast.AttributeSpec] {
	return func(p *parser.Parser) (*ast.AttributeSpec, *diagnostic.Diagnostic) {
		var spec ast.AttributeSpec

		spec.Selector = parser.Try(p, Selector())
		if spec.Selector == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute spec",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute selector"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		spec.Ruleset = parser.Must(p, Ruleset())
		return &spec, nil
	}
}

func Ruleset() parser.Func[*ast.AttributeRuleset] {
	return func(p *parser.Parser) (*ast.AttributeRuleset, *diagnostic.Diagnostic) {
		var rs ast.AttributeRuleset

		rs.LBrace = parser.TryRuneAt(p, '{')
		if rs.LBrace == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute ruleset",
				Primary: quickanno.Expected(p, p.Pos(), "an opening brace"),
			}
		}

		rs.Rules = make([]*ast.AttributeRule, 0, 64)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			r := parser.TryOptional(p, Rule(), nil)
			if r == nil {
				break
			}
			parser.MustSkip(p, comment.AndMustEOS())
			rs.Rules = append(rs.Rules, r)
		}
		rs.Rules = slices.Clip(rs.Rules)

		rs.RBrace = parser.TryRuneAt(p, '}')
		if rs.RBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute ruleset: unclosed brace",
				Primary: quickanno.Expected(p, *rs.LBrace, "expected a `}` for the opening `{` here"),
			})
		}

		return &rs, nil
	}
}

func Rule() parser.Func[*ast.AttributeRule] {
	return func(p *parser.Parser) (*ast.AttributeRule, *diagnostic.Diagnostic) {
		var r ast.AttributeRule

		r.Selector = parser.Try(p, ElementSelector())
		if r.Selector == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute rule",
				Primary: quickanno.Expected(p, p.Pos(), "an element selector"),
			}
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		r.Type = parser.Must(p, TypeName())
		return &r, nil
	}
}

func Selector() parser.Func[ast.AttributeSelector] {
	return func(p *parser.Parser) (ast.AttributeSelector, *diagnostic.Diagnostic) {
		if bs := parser.Try(p, BasicSelector()); bs != nil {
			return bs, nil
		} else if rs := parser.Try(p, RegexpSelector()); rs != nil {
			return rs, nil
		}
		return nil, &diagnostic.Diagnostic{
			Message: "missing attribute selector",
			Primary: quickanno.Expected(p, p.Pos(), "an attribute selector"),
		}
	}
}

func BasicSelector() parser.Func[*ast.BasicAttributeSelector] {
	return func(p *parser.Parser) (*ast.BasicAttributeSelector, *diagnostic.Diagnostic) {
		var s ast.BasicAttributeSelector
		s.Position = p.PosPtr()

		name := parser.Try(p, Name())
		if name == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing basic attribute selector",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute name"),
			}
		}
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

		return &s, nil
	}
}

func RegexpSelector() parser.Func[*ast.RegexpAttributeSelector] {
	return func(p *parser.Parser) (*ast.RegexpAttributeSelector, *diagnostic.Diagnostic) {
		var s ast.RegexpAttributeSelector

		s.Regexp = parser.TryTokenAt(p, "'regexp")
		if s.Regexp == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing regexp attribute selector",
				Primary:  quickanno.Expected(p, p.Pos(), "a regexp attribute selector"),
				Examples: []diagnostic.Example{{Example: "'regexp(`hx-\\d{3}`)"}},
			}
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		s.LParen = parser.TryRuneAt(p, '(')
		if s.LParen == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "regexp attribute selector: missing arguments",
				Primary:  quickanno.Expected(p, p.Pos(), "an opening parenthesis"),
				Examples: []diagnostic.Example{{Example: "'regexp(`hx-\\d{3}`)"}},
			}
		}
		parser.TrySkip(p, whitespace.Any())

		s.Raw = parser.Try(p, golang.StringLit())
		if s.Raw == nil || s.Raw.Contents == "" {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "regexp attribute selector: missing regexp",
				Primary:  quickanno.Expected(p, p.Pos(), "a string containing a regular expression"),
				Examples: []diagnostic.Example{{Example: "'regexp(`hx-\\d{3}`)"}},
			})
		}

		var err error
		s.Compiled, err = regexp.Compile(s.Raw.Unquote())
		if err != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "invalid regular expression: " + err.Error(),
				Primary: quickanno.Expected(p, s.Raw.Start(), "a valid regular expression"),
			})
		}

		parser.TrySkip(p, whitespace.Any())
		s.RParen = parser.TryRuneAt(p, ')')
		if s.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "regexp attribute selector: missing closing parenthesis",
				Primary: quickanno.Expected(p, *s.LParen, "a closing parenthesis"),
			})
		}

		return &s, nil
	}
}

func ElementSelector() parser.Func[ast.ElementSelector] {
	return func(p *parser.Parser) (ast.ElementSelector, *diagnostic.Diagnostic) {
		if ws := parser.Try(p, WildcardElementSelector()); ws != nil {
			return ws, nil
		} else if ls := parser.Try(p, ListElementSelector()); ls != nil {
			return ls, nil
		}
		return nil, &diagnostic.Diagnostic{
			Message: "missing element selector",
			Primary: quickanno.Expected(p, p.Pos(), "an element selector"),
			Examples: []diagnostic.Example{
				{Example: "*", Title: "wildcard selector"},
				{Example: "div, span", Title: "list selector"},
			},
		}
	}
}

func WildcardElementSelector() parser.Func[*ast.WildcardElementSelector] {
	return func(p *parser.Parser) (*ast.WildcardElementSelector, *diagnostic.Diagnostic) {
		var s ast.WildcardElementSelector

		s.Asterisk = parser.TryRuneAt(p, '*')
		if s.Asterisk == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing wildcard element selector",
				Primary: quickanno.Expected(p, p.Pos(), "an asterisk"),
			}
		}
		return &s, nil
	}
}

func ListElementSelector() parser.Func[*ast.ListElementSelector] {
	return func(p *parser.Parser) (*ast.ListElementSelector, *diagnostic.Diagnostic) {
		var s ast.ListElementSelector

		var err *diagnostic.Diagnostic
		s.Elements, err = parser.TryErr(p, list.CommaList("element name", "element names", elementName()))
		if err != nil {
			return nil, err
		}

		return &s, nil
	}
}

func elementName() parser.Func[*ast.ElementName] {
	return func(p *parser.Parser) (*ast.ElementName, *diagnostic.Diagnostic) {
		var n ast.ElementName
		n.Position = p.PosPtr()

		n.Name = parser.Try(p, html.TagName())
		if n.Name == "" {
			return nil, &diagnostic.Diagnostic{
				Message: "missing element name",
				Primary: quickanno.Expected(p, *n.Position, "an html element name"),
			}
		}

		return &n, nil
	}
}
