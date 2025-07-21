package element

import (
	"fmt"

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
	return func(p *parser.Parser) (*ast.Doctype, *diagnostic.Diagnostic) {
		var d ast.Doctype

		d.Doctype = parser.TryTokenAt(p, "!doctype")
		if d.Doctype == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing doctype",
				Primary: quickanno.Expected(p, p.Pos(), "a doctype"),
			}
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		args := parser.Try(p, argument.Arguments())
		if args == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "doctype: missing html attribute",
				Primary: quickanno.Expected(p, p.Pos(), "`(html)`"),
				Examples: []diagnostic.Example{
					{Example: "`!doctype(html)`"},
				},
			})
			return &d, nil
		}
		d.LParen, d.RParen = args.LParen, args.RParen
		switch {
		case len(args.Args) == 0:
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "doctype: missing html attribute",
				Primary: quickanno.Expected(p, *args.LParen, "an html attribute"),
				Examples: []diagnostic.Example{
					{Example: "`!doctype(html)`"},
				},
			})
		case len(args.Args) == 1:
			attr, ok := args.Args[0].(*ast.NamedAttribute)
			if !ok {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "doctype: invalid html attribute",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, args.Args[0].Start(), args.Args[0].End(), "expected `html`, not this"),
					},
					Examples: []diagnostic.Example{{Example: "`!doctype(html)`"}},
				})
			} else {
				d.HTML = new(ast.Position)
				*d.HTML = attr.Name.Start()
				if attr.Name.Package != nil || attr.Name.Name.Name != "html" {
					d.HTML = nil
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "doctype: missing html attribute",
						Primary: []diagnostic.Annotation{
							anno.Range(p.File, attr.Name.Start(), attr.Name.End(), "expected `html`, not this"),
						},
						Examples: []diagnostic.Example{{Example: "`!doctype(html)`"}},
					})
				} else if attr.EqualSign != nil {
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
					anno.Range(p.File, args.Args[0].Start(), args.Args[len(args.Args)-1].End(), "only a single `html` attribute"),
				},
				Examples: []diagnostic.Example{{Example: "`!doctype(html)`"}},
			})
		}
		d.RParen = args.RParen

		return &d, nil
	}
}

func Element() parser.Func[*ast.Element] {
	return func(p *parser.Parser) (*ast.Element, *diagnostic.Diagnostic) {
		var e ast.Element

		e.Header = parser.Try(p, Header())
		if e.Header == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing element",
				Primary: quickanno.Expected(p, p.Pos(), "an element"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		e.Body = parser.Try(p, body.Body())
		return &e, nil
	}
}

func Header() parser.Func[*ast.ElementHeader] {
	return func(p *parser.Parser) (*ast.ElementHeader, *diagnostic.Diagnostic) {
		var h ast.ElementHeader

		h.Name = parser.Try(p, Reference())
		if h.Name == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing element header",
				Primary: quickanno.Expected(p, p.Pos(), "an element name"),
			}
		}
		if !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		h.Attributes = parser.Try(p, argument.Arguments())
		return &h, nil
	}
}

func Reference() parser.Func[*ast.ElementReference] {
	return func(p *parser.Parser) (*ast.ElementReference, *diagnostic.Diagnostic) {
		var ref ast.ElementReference

		state := p.CloneState()

		ref.Package = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		ref.Dot = parser.TryOptionalRuneAt(p, '.', comment.OrAnyWhitespace())
		if ref.Dot == nil {
			ref.Package = nil
			p.RestoreState(state)
		} else if ref.Package == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute reference: missing package name",
				Primary: quickanno.Expected(p, p.Pos(), "a package name before the `.`"),
			})
		}

		var err *diagnostic.Diagnostic
		ref.Name, err = parser.TryErr(p, Name())
		if err != nil {
			return nil, err
		}
		return &ref, nil
	}
}

func Name() parser.Func[*ast.ElementName] {
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

func Raw() parser.Func[*ast.RawElement] {
	return func(p *parser.Parser) (*ast.RawElement, *diagnostic.Diagnostic) {
		var e ast.RawElement

		e.Raw = parser.TryTokenAt(p, "!raw")
		if e.Raw == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing raw element",
				Primary: quickanno.Expected(p, p.Pos(), "a raw element"),
			}
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		args := parser.TryOptional(p, argument.Arguments(), comment.OrHorizontalWhitespace())
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

		return &e, nil
	}
}

func And() parser.Func[*ast.And] {
	return func(p *parser.Parser) (*ast.And, *diagnostic.Diagnostic) {
		var a ast.And

		a.And = parser.TryTokenAt(p, "&")
		if a.And == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing &-attributes",
				Primary: quickanno.Expected(p, p.Pos(), "&"),
			}
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		a.Attributes = parser.Try(p, argument.Arguments())
		if a.Attributes == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "&-attributes: missing attributes",
				Primary: quickanno.Expected(p, p.Pos(), "attributes"),
				Examples: []diagnostic.Example{
					{Example: "`&(aria-label=\"woof\")`"},
				},
			})
		}

		return &a, nil
	}
}
