package attribute

import (
	"regexp"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func Definition() parser.Func[*ast.AttributeDefinition] {
	return func(p *parser.Parser) (*ast.AttributeDefinition, *fancyerr.Error) {
		def := &ast.AttributeDefinition{Position: p.Pos()}

		if !parser.TryToken(p, "attr") || !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing attribute definition",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute definition"),
			}
		}

		defer parser.MustSkip(p, comment.AndMustEOS())

		beforePrefix := p.CloneState()

		// technically '(' would be a valid attribute name, so check that we
		// don't accidentally consume a '(' as a prefix here
		if !parser.MatchesAnyRune(p, '(') {
			var ok bool
			def.Prefix, ok = parser.TryOptionalOk(p, Name())
			if ok {
				parser.TrySkipOk(p, comment.OrAnyWhitespace())
			}
		}

		// check if this is a single spec
		if !parser.MatchesAnyRune(p, '(') {
			if parser.MatchesAnyRune(p, '{') {
				// our prefix is actually a single spec
				def.Prefix = nil
				p.RestoreState(beforePrefix)
			}

			s, err := parser.Try(p, Spec())
			if err != nil {
				p.CaptureError(err)
			} else {
				def.Specs = []*ast.AttributeSpec{s}
			}
			return def, nil
		}

		def.LParen = p.PosPtr()
		parser.TryRune(p, '(')

		def.Specs = make([]*ast.AttributeSpec, 0, 64)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			s, ok := parser.TryOk(p, Spec())
			if !ok {
				break
			}
			def.Specs = append(def.Specs, s)
		}
		def.Specs = slices.Clip(def.Specs)

		parser.TrySkip(p, comment.OrAnyWhitespace())
		def.RParen = p.PosPtr()
		if !parser.TryRune(p, ')') {
			def.RParen = nil
			p.CaptureError(&fancyerr.Error{
				Message: "unclosed attribute definition",
				Primary: quickanno.Expected(p, *def.LParen, "expected a `)` for the opening `(` here"),
			})
		}
		return def, nil
	}
}

func Spec() parser.Func[*ast.AttributeSpec] {
	return func(p *parser.Parser) (*ast.AttributeSpec, *fancyerr.Error) {
		sel, ok := parser.TryOk(p, Selector())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing attribute spec",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute selector"),
			}
		}

		spec := &ast.AttributeSpec{Name: sel}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		spec.Ruleset = parser.Must(p, Ruleset())

		return spec, nil
	}
}

func Ruleset() parser.Func[*ast.AttributeRuleset] {
	return func(p *parser.Parser) (*ast.AttributeRuleset, *fancyerr.Error) {
		rs := &ast.AttributeRuleset{LBrace: p.Pos()}

		if !parser.TryRune(p, '{') {
			return nil, &fancyerr.Error{
				Message: "missing attribute ruleset",
				Primary: quickanno.Expected(p, rs.LBrace, "an opening brace"),
			}
		}

		rs.Rules = make([]*ast.AttributeRule, 0, 64)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			r, ok := parser.TryOk(p, Rule())
			if !ok {
				break
			}
			rs.Rules = append(rs.Rules, r)

			parser.MustSkip(p, comment.AndMustEOS())
		}
		rs.Rules = slices.Clip(rs.Rules)

		parser.TrySkip(p, comment.OrAnyWhitespace())

		rs.RBrace = p.PosPtr()
		if !parser.TryRune(p, '}') {
			rs.RBrace = nil
			p.CaptureError(&fancyerr.Error{
				Message: "unclosed attribute ruleset",
				Primary: quickanno.Expected(p, rs.LBrace, "expected a `}` for the opening `{` here"),
			})
		}

		return rs, nil
	}
}

func Rule() parser.Func[*ast.AttributeRule] {
	return func(p *parser.Parser) (*ast.AttributeRule, *fancyerr.Error) {
		sel, ok := parser.TryOk(p, ElementSelector())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing attribute rule",
				Primary: quickanno.Expected(p, p.Pos(), "an element selector"),
			}
		}

		r := &ast.AttributeRule{Selector: sel}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		r.Type = parser.Must(p, TypeName())

		return r, nil
	}
}

func Selector() parser.Func[ast.AttributeSelector] {
	return func(p *parser.Parser) (ast.AttributeSelector, *fancyerr.Error) {
		if bs, ok := parser.TryOk(p, BasicSelector()); ok {
			return bs, nil
		} else if rs, ok := parser.TryOk(p, RegexpSelector()); ok {
			return rs, nil
		}
		return nil, &fancyerr.Error{
			Message: "missing attribute selector",
			Primary: quickanno.Expected(p, p.Pos(), "an attribute selector"),
		}
	}
}

func BasicSelector() parser.Func[*ast.BasicAttributeSelector] {
	return func(p *parser.Parser) (*ast.BasicAttributeSelector, *fancyerr.Error) {
		s := &ast.BasicAttributeSelector{Position: p.Pos()}

		name, ok := parser.TryOk(p, Name())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing basic attribute selector",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute name"),
			}
		}
		s.Name = name.Name

		s.Wildcard = strings.HasSuffix(s.Name, "*")
		if s.Wildcard {
			s.Name = s.Name[:len(s.Name)-1]
			if s.Name == "" {
				p.CaptureError(&fancyerr.Error{
					Message:     "basic attribute selector: only wildcard",
					Primary:     quickanno.Expected(p, s.Position, "an attribute name before the wildcard"),
					Explanation: "A basic attribute selector mustn't consist solely of a wildcard.",
				})
			}
		}

		return s, nil
	}
}

func RegexpSelector() parser.Func[*ast.RegexpAttributeSelector] {
	return func(p *parser.Parser) (*ast.RegexpAttributeSelector, *fancyerr.Error) {
		s := &ast.RegexpAttributeSelector{Position: p.Pos()}

		if !parser.TryToken(p, "'regexp") {
			return nil, &fancyerr.Error{
				Message: "missing regexp attribute selector",
				Primary: quickanno.Expected(p, p.Pos(), "'regexp"),
			}
		}

		hasWS := parser.TrySkipOk(p, comment.OrHorizontalWhitespace())
		s.LParen = p.PosPtr()
		if !parser.TryRune(p, '(') {
			s.LParen = nil
			err := &fancyerr.Error{
				Message: "regexp attribute selector: missing opening parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "an opening parenthesis"),
			}
			if hasWS {
				p.CaptureError(err)
				return s, nil
			}
			return nil, err
		}

		parser.TrySkip(p, whitespace.Any())

		var ok bool
		s.Raw, ok = parser.TryOk(p, golang.StringLit())
		if !ok || s.Raw.Contents == "" {
			p.CaptureError(&fancyerr.Error{
				Message:  "regexp attribute selector: missing regexp",
				Primary:  quickanno.Expected(p, p.Pos(), "a string containing a regular expression"),
				Examples: []fancyerr.Example{{Example: "'regexp(`hx-\\d{3}`)"}},
			})
		}

		var err error
		s.Regexp, err = regexp.Compile(s.Raw.Unquote())
		if err != nil {
			p.CaptureError(&fancyerr.Error{
				Message: "invalid regular expression: " + err.Error(),
				Primary: quickanno.Expected(p, s.Raw.Pos(), "a valid regular expression"),
			})
		}

		parser.TrySkip(p, whitespace.Any())
		s.RParen = p.PosPtr()
		if !parser.TryRune(p, ')') {
			s.RParen = nil
			p.CaptureError(&fancyerr.Error{
				Message: "regexp attribute selector: missing closing parenthesis",
				Primary: quickanno.Expected(p, *s.LParen, "a closing parenthesis"),
			})
		}

		return s, nil
	}
}

func ElementSelector() parser.Func[ast.ElementSelector] {
	return func(p *parser.Parser) (ast.ElementSelector, *fancyerr.Error) {
		if ws, ok := parser.TryOk(p, WildcardElementSelector()); ok {
			return ws, nil
		} else if ls, ok := parser.TryOk(p, ListElementSelector()); ok {
			return ls, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing element selector",
			Primary: quickanno.Expected(p, p.Pos(), "an element selector"),
			Examples: []fancyerr.Example{
				{Example: "*", Title: "wildcard selector"},
				{Example: "div, span", Title: "list selector"},
			},
		}
	}
}

func WildcardElementSelector() parser.Func[*ast.WildcardElementSelector] {
	return func(p *parser.Parser) (*ast.WildcardElementSelector, *fancyerr.Error) {
		s := &ast.WildcardElementSelector{Asterisk: p.Pos()}
		if !parser.TryRune(p, '*') {
			return nil, &fancyerr.Error{
				Message: "missing wildcard element selector",
				Primary: quickanno.Expected(p, s.Asterisk, "an asterisk"),
			}
		}
		return s, nil
	}
}

func ListElementSelector() parser.Func[*ast.ListElementSelector] {
	return func(p *parser.Parser) (*ast.ListElementSelector, *fancyerr.Error) {
		s := &ast.ListElementSelector{Elements: make([]*ast.ListElementSelectorItem, 0, 12)}

		var err *fancyerr.Error
		s.Elements, err = parser.Try(p, list.CommaList("element name", "element names", ListElementSelectorItem()))
		if err != nil {
			return nil, err
		}

		return s, nil
	}
}

func ListElementSelectorItem() parser.Func[*ast.ListElementSelectorItem] {
	return func(p *parser.Parser) (*ast.ListElementSelectorItem, *fancyerr.Error) {
		itm := &ast.ListElementSelectorItem{Position: p.Pos()}

		var ok bool
		itm.Name, ok = parser.TryOk(p, html.TagName())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing element name",
				Primary: quickanno.Expected(p, itm.Position, "an html element name"),
			}
		}

		return itm, nil
	}
}
