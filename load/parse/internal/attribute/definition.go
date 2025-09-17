package attribute

import (
	"regexp"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
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
		attr := parser.TryKeywordAt(p, "attr")
		if attr == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		def := parser.TryInOrder(p,
			definitionList(attr), singleDefinitionWithPrefix(attr), singleDefinitionWithoutPrefix(attr))
		if def != nil {
			return def
		}

		p.CaptureError(&diagnostic.Diagnostic{
			Message: "attribute definition: missing attribute name",
			Primary: quickanno.Expected(p, p.Pos(), "an attribute name or an opening `(`"),
			Examples: []diagnostic.Example{
				{Example: "`attr woof { ... }`", Title: "single definition"},
				{Example: "`attr ( ... )`", Title: "definition list"},
			},
		})
		return &ast.AttributeDefinition{Attr: attr}
	}
}

func definitionList(attr *ast.Position) parser.Func[*ast.AttributeDefinition] {
	return func(p *parser.Parser) *ast.AttributeDefinition {
		// technically '(' would be a valid attribute name, so check that we
		// don't accidentally consume a '(' as a prefix here
		var prefix *ast.AttributeName
		if !parser.MatchesAnyRune(p, '(') {
			prefix = parser.TryOptional(p, Name(), comment.OrHorizontalWhitespace())
		}

		lParen := parser.TryRuneAt(p, '(')
		if lParen == nil {
			return nil
		}

		var def ast.AttributeDefinition
		def.Attr = attr
		def.Prefix = prefix
		def.LParen = lParen

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

func singleDefinitionWithPrefix(attr *ast.Position) parser.Func[*ast.AttributeDefinition] {
	return func(p *parser.Parser) *ast.AttributeDefinition {
		prefix := parser.TryOptional(p, Name(), comment.OrHorizontalWhitespace())
		if prefix == nil {
			return nil
		}

		if parser.MatchesAnyRune(p, '{') {
			// prefix was actually start of the spec
			return nil
		}

		spec := parser.Try(p, Spec())
		if spec == nil {
			return nil
		}

		return &ast.AttributeDefinition{
			Attr:   attr,
			Prefix: prefix,
			Specs:  []*ast.AttributeSpec{spec},
		}
	}
}

func singleDefinitionWithoutPrefix(attr *ast.Position) parser.Func[*ast.AttributeDefinition] {
	return func(p *parser.Parser) *ast.AttributeDefinition {
		spec := parser.Try(p, Spec())
		if spec == nil {
			return nil
		}

		return &ast.AttributeDefinition{
			Attr:  attr,
			Specs: []*ast.AttributeSpec{spec},
		}
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

		s.CanonicalName = canonicalize(s.Name)
		return &s
	}
}

func canonicalize(name string) string {
	var b strings.Builder
	for i, r := range name {
		if r < 'A' || r > 'Z' {
			continue
		}

		b.Grow(len(name))
		b.WriteString(name[:i])
		b.WriteRune((r - 'A') + 'a')
		for _, r := range name[i+1:] {
			if r >= 'A' && r <= 'Z' {
				b.WriteRune((r - 'A') + 'a')
			} else {
				b.WriteRune(r)
			}
		}
	}

	if b.Len() == 0 {
		return name
	}
	return b.String()
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
			expr := s.Raw.Unquote()
			if strings.HasPrefix(expr, "^") {
				caretPos := s.Raw.Start()
				caretPos.Col += len(`"`)
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "regexp attribute selector: unnecessary start anchor",
					Primary: []diagnostic.Annotation{
						anno.Position(p.File, caretPos, "remove this start anchor"),
					},
					Explanation: "Matches against a regular expression selector always match the entire attribute name anyway, " +
						"so there is no point in adding this start anchor.\n" +
						"Remove it to avoid confusion.",
					Hints: []diagnostic.Hint{
						{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
					},
				})
			} else {
				expr = "^" + expr
			}
			if strings.HasSuffix(expr, "$") {
				dollarPos := s.Raw.End()
				dollarPos.Col -= len(`"`)
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "regexp attribute selector: unnecessary end anchor",
					Primary: []diagnostic.Annotation{
						anno.Position(p.File, dollarPos, "remove this end anchor"),
					},
					Explanation: "Matches against a regular expression selector always match the entire attribute name anyway, " +
						"so there is no point in adding this end anchor.\n" +
						"Remove it to avoid confusion.",
					Hints: []diagnostic.Hint{
						{Hint: "The formatter (`corgi fmt`) can automatically fix this error."},
					},
				})
			} else {
				expr = expr + "$"
			}

			var err error
			s.Compiled, err = regexp.Compile(expr)
			if err != nil {
				_, origErr := regexp.Compile(s.Raw.Unquote()) // try without anchors to get a better error message
				if origErr != nil {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "regexp attribute selector: invalid regular expression",
						Primary: []diagnostic.Annotation{anno.Node(p.File, s.Raw, err.Error())},
					})
				} else {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "regexp attribute selector: invalid regular expression: could not apply anchors",
						Primary: []diagnostic.Annotation{anno.Node(p.File, s.Raw, err.Error())},
						Explanation: "The anchors (`^` and `$`) are automatically added to the regular expression, " +
							"to enforce full matches.\n" +
							"When trying to compile the regular expression with anchors, it fails, but without anchors it succeeds.",
					})
				}
			} else {
				prefix, _ := s.Compiled.LiteralPrefix()
				for _, r := range prefix {
					if r >= 'A' && r <= 'Z' {
						p.CaptureError(&diagnostic.Diagnostic{
							Message: "regexp attribute selector: expression is not in canonical form",
							Primary: []diagnostic.Annotation{
								anno.Node(p.File, s.Raw, "contains uppercase ASCII letters"),
							},
							Explanation: "HTML attribute names are ASCII-case-insensitive, meaning that uppercase and " +
								"lowercase ASCII letters are considered equivalent." +
								"To satisfy this interchangeability, attribute names are always lowercased before matching, " +
								"so regular expression selectors must always accept the lowercase variant of an attribute name " +
								"to be able to make a match.\n" +
								"This restriction does not apply to non-ASCII letters, which are case-sensitive in HTML.",
							Hints: []diagnostic.Hint{
								{Hint: "Only use lowercase ASCII letters in the regular expression."},
							},
						})
						break
					}
				}
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
		names := parser.Try(p, list.CommaList("list element-selector", "element name", "element names", elementReference))
		if len(names) == 0 {
			return nil
		}

		return &ast.ListElementSelector{List: names}
	}
}

var elementReference parser.Func[*ast.ElementReference]

func SetElementReference(f parser.Func[*ast.ElementReference]) {
	elementReference = f
}
