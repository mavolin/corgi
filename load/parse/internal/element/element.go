package element

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/argument"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Doctype() parser.Func[*ast.Doctype] {
	return func(p *parser.Parser) *ast.Doctype {
		doctype := parser.TryTokenAt(p, "!doctype")
		if doctype == nil {
			return nil
		}

		var d ast.Doctype
		d.Doctype = doctype
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		argsStart := p.Pos()
		args := parser.Try(p, argument.Arguments("doctype"))
		argsEnd := p.Pos()
		if args == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "doctype: missing html attribute",
				Primary: quickanno.Expected(p, p.Pos(), "`(html)`"),
				Examples: []diagnostic.Example{
					{Example: "`!doctype(html)`"},
				},
			})
			return &d
		}
		d.LParen, d.RParen = args.LParen, args.RParen
		switch {
		case len(args.List) == 0:
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "doctype: missing html attribute",
				Primary: quickanno.Expected(p, *args.LParen, "an html attribute"),
				Examples: []diagnostic.Example{
					{Example: "`!doctype(html)`"},
				},
			})
		case len(args.List) == 1:
			attr, _ := args.List[0].(*ast.NamedAttribute)
			if attr == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "doctype: invalid html attribute",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, argsStart, argsEnd, "expected `html`, not this"),
					},
					Examples: []diagnostic.Example{{Example: "`!doctype(html)`"}},
				})
			} else {
				if attr.Name != nil && attr.Name.Name != nil && attr.Name.Package == nil && attr.Name.Name.Name == "html" {
					d.HTML = new(ast.Position)
					*d.HTML = attr.Name.Start()
				} else {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "doctype: missing html attribute",
						Primary: []diagnostic.Annotation{
							anno.Node(p.File, attr, "expected `html`, not this"),
						},
						Examples: []diagnostic.Example{{Example: "`!doctype(html)`"}},
					})
				}

				if attr.EqualSign != nil || attr.Value != nil {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "doctype: html attribute: unexpected value",
						Primary: []diagnostic.Annotation{
							anno.Range(p.File, *attr.EqualSign, attr.End(), "remove this invalid value"),
						},
						Examples: []diagnostic.Example{{Example: "`!doctype(html)`"}},
					})
				}
			}
		default:
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "doctype: too many attributes",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, argsStart, argsEnd, "only a single `html` attribute"),
				},
				Examples: []diagnostic.Example{{Example: "`!doctype(html)`"}},
			})
		}
		d.RParen = args.RParen

		return &d
	}
}

func Element() parser.Func[*ast.Element] {
	return func(p *parser.Parser) *ast.Element {
		header := parser.Try(p, Header())
		if header == nil {
			return nil
		}

		var e ast.Element
		e.Header = header

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		e.Body = parser.Try(p, body.Body())

		// Only match, if we have a name, and either arguments, a body or have
		// reached the EOS.
		// Otherwise, we can't be sure that this is actually an element, since
		// we might've matched the start of an assignment.
		if e.Header.Attributes == nil && e.Body == nil && !parser.Matches(p, comment.AndEOS()) {
			return nil
		}

		return &e
	}
}

func Header() parser.Func[*ast.ElementHeader] {
	return func(p *parser.Parser) *ast.ElementHeader {
		name := parser.Try(p, Reference())
		if name == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var h ast.ElementHeader
		h.Name = name

		h.Attributes = parser.Try(p, argument.Arguments("element"))
		return &h
	}
}

func Reference() parser.Func[*ast.ElementReference] {
	return func(p *parser.Parser) *ast.ElementReference {
		return parser.TryInOrder(p, qualifiedReference(), unqualifiedReference())
	}
}

func qualifiedReference() parser.Func[*ast.ElementReference] {
	return func(p *parser.Parser) *ast.ElementReference {
		pkg := parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		if pkg == nil {
			return nil
		}

		dot := parser.TryOptionalRuneAt(p, '.', comment.OrAnyWhitespace())
		if dot == nil {
			return nil
		}

		var ref ast.ElementReference
		ref.Package = pkg
		ref.Dot = dot

		ref.Name = parser.Try(p, Name())
		if ref.Name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "element reference: missing element name",
				Primary: quickanno.Expected(p, p.Pos(), "an element name after the `.`"),
			})
		}

		return &ref
	}
}

func unqualifiedReference() parser.Func[*ast.ElementReference] {
	return func(p *parser.Parser) *ast.ElementReference {
		name := parser.Try(p, Name())
		if name == nil {
			return nil
		}

		return &ast.ElementReference{Name: name}
	}
}

func Name() parser.Func[*ast.ElementName] {
	return func(p *parser.Parser) *ast.ElementName {
		name := parser.TokenWhile(p, func() bool {
			if !parser.MatchesRunePredicate(p, html.AttributeNameRune) {
				return false
			}

			return !parser.MatchesAnyRune(p, '.', ',', '\r', ';', '(', '{', '[', '}', ']', ')', '_')
		})
		if name == "" {
			return nil
		}

		pos := p.Pos()
		pos.Col -= ast.Col(utf8.RuneCountInString(name)) //nolint:gosec

		return &ast.ElementName{
			Name:          name,
			CanonicalName: canonicalize(name),
			Position:      &pos,
		}
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
		b.WriteRune('a' + (r - 'A'))
		for _, r := range name[i+1:] {
			if r >= 'A' && r <= 'Z' {
				b.WriteRune('a' + (r - 'A'))
			} else {
				b.WriteRune(r)
			}
		}
		break
	}

	if b.Len() == 0 {
		return name
	}
	return b.String()
}

func Raw() parser.Func[*ast.RawElement] {
	return func(p *parser.Parser) *ast.RawElement {
		raw := parser.TryTokenAt(p, "!raw")
		if raw == nil {
			return nil
		}

		var e ast.RawElement
		e.Raw = raw

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		args := parser.TryOptional(p, argument.Arguments("raw element"), comment.OrHorizontalWhitespace())
		if args != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "raw element: unexpected attributes",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, args.Start(), args.End(), "remove these attributes"),
				},
				Explanation: "Only the body of a `!raw` element is rendered, but not the `!raw` element itself. " +
					"Thus, there is no point in placing attributes on a `!raw` element.",
				Docs: "raw-element",
			})
		}

		e.Body = parser.TryOptional(p, body.VerbatimBracketText(), nil)
		if e.Body == nil {
			b := parser.Try(p, body.Body())
			if b != nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "raw element: non-bracket-text body",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, b.Start(), b.End(), fmt.Sprintf("expected bracket text, not %T", b)),
					},
					Explanation: "Because of their nature, `!raw` elements only support bracket text bodies.`",
					Docs:        "raw-element",
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "raw element: missing body",
					Primary: quickanno.Expected(p, *e.Raw, "brace text"),
				})
			}
		}

		return &e
	}
}

func And() parser.Func[*ast.And] {
	return func(p *parser.Parser) *ast.And {
		and := parser.TryTokenAt(p, "&")
		if and == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var a ast.And
		a.And = and

		a.Attributes = parser.Try(p, argument.Arguments("&"))
		if a.Attributes == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "&-attributes: missing attributes",
				Primary: quickanno.Expected(p, p.Pos(), "attributes"),
				Examples: []diagnostic.Example{
					{Example: "`&(aria-label=\"woof\")`"},
				},
			})
		}

		return &a
	}
}
