package element

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/argument"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/html"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Doctype() parser.Func[*ast.Doctype] {
	return func(p *parser.Parser) (*ast.Doctype, *fancyerr.Error) {
		d := &ast.Doctype{Doctype: p.Pos()}
		if !parser.TryToken(p, "!doctype") {
			return nil, &fancyerr.Error{
				Message: "missing doctype",
				Primary: quickanno.Expected(p, p.Pos(), "a doctype"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		args, ok := parser.TryOk(p, argument.Arguments())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "doctype: missing html attribute",
				Primary: quickanno.Expected(p, p.Pos(), "`(html)`"),
				Examples: []fancyerr.Example{
					{Example: "`!doctype(html)`"},
				},
			})
			return d, nil
		}
		d.LParen = &args.LParen
		if len(args.Args) == 0 {
			p.CaptureError(&fancyerr.Error{
				Message: "doctype: missing html attribute",
				Primary: quickanno.Expected(p, args.LParen, "an html attribute"),
			})
		} else if len(args.Args) == 1 {
			attr, ok := args.Args[0].(*ast.NamedAttribute)
			if !ok {
				p.CaptureError(&fancyerr.Error{
					Message: "doctype: invalid html attribute",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, args.Args[0].Pos(), args.Args[0].End(), "expected `html`, not this"),
					},
					Examples: []fancyerr.Example{
						{Example: "`!doctype(html)`"},
					},
				})
			} else {
				d.HTML = &attr.Name.Position
				if attr.Name.Name != "html" {
					d.HTML = nil
					p.CaptureError(&fancyerr.Error{
						Message: "doctype: missing html attribute",
						Primary: []fancyerr.Annotation{
							anno.Range(p.File, attr.Name.Pos(), attr.Name.End(), "expected `html`, not this"),
						},
						Examples: []fancyerr.Example{
							{Example: "`!doctype(html)`"},
						},
					})
				} else if attr.EqualSign != nil {
					p.CaptureError(&fancyerr.Error{
						Message: "doctype: html attribute: unexpected value",
						Primary: []fancyerr.Annotation{
							anno.Range(p.File, *attr.EqualSign, attr.End(), "remove this invalid value"),
						},
						Examples: []fancyerr.Example{
							{Example: "`!doctype(html)`"},
						},
					})
				}
			}
		} else {
			p.CaptureError(&fancyerr.Error{
				Message: "doctype: too many attributes",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, args.Args[0].Pos(), args.Args[len(args.Args)-1].End(), "only a single `html` attribute"),
				},
				Examples: []fancyerr.Example{
					{Example: "`!doctype(html)`"},
				},
			})
		}
		d.RParen = args.RParen

		return d, nil
	}
}

func Element() parser.Func[*ast.Element] {
	return func(p *parser.Parser) (*ast.Element, *fancyerr.Error) {
		h, ok := parser.TryOk(p, Header())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing element",
				Primary: quickanno.Expected(p, p.Pos(), "an element"),
			}
		}

		e := &ast.Element{Header: *h}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		e.Body, _ = parser.Try(p, body.Body())

		return e, nil
	}
}

func Header() parser.Func[*ast.ElementHeader] {
	return func(p *parser.Parser) (*ast.ElementHeader, *fancyerr.Error) {
		name, ok := parser.TryOk(p, Name())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing element header",
				Primary: quickanno.Expected(p, p.Pos(), "an element name"),
			}
		}

		h := &ast.ElementHeader{Name: *name}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		h.Attributes, _ = parser.Try(p, argument.Arguments())
		return h, nil
	}
}

func Name() parser.Func[*ast.ElementName] {
	return func(p *parser.Parser) (*ast.ElementName, *fancyerr.Error) {
		n := &ast.ElementName{Position: p.Pos()}
		n.Name, _ = parser.Try(p, html.TagName())
		if n.Name == "" {
			return nil, &fancyerr.Error{
				Message: "missing element name",
				Primary: quickanno.Expected(p, n.Position, "an html element name"),
			}
		}

		return n, nil
	}
}

func Raw() parser.Func[*ast.RawElement] {
	return func(p *parser.Parser) (*ast.RawElement, *fancyerr.Error) {
		e := &ast.RawElement{Raw: p.Pos()}
		if !parser.TryToken(p, "!raw") {
			return nil, &fancyerr.Error{
				Message: "missing raw element",
				Primary: quickanno.Expected(p, p.Pos(), "a raw element"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		args, _ := parser.TryOptional(p, argument.Arguments())
		if args != nil {
			p.CaptureError(&fancyerr.Error{
				Message: "raw element: unexpected attributes",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, args.Pos(), args.End(), "remove these attributes"),
				},
				Explanation: "Only the body of a `!raw` element is rendered, but not the `!raw` element itself. " +
					"Thus, there is no point in placing attributes on a `!raw` element.",
				Docs: "!raw-element",
			})
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var ok bool
		e.Body, ok = parser.TryOptionalOk(p, body.BracketText())
		if !ok {
			b, _ := parser.Try(p, body.Body())
			if b != nil {
				p.CaptureError(&fancyerr.Error{
					Message: "raw element: non-bracket-text body",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, b.Pos(), b.End(), "expected bracket text"),
					},
					Explanation: "Because of their nature, `!raw` elements only support bracket text bodies.`",
					Docs:        "!raw-element",
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "raw element: missing body",
					Primary: quickanno.Expected(p, e.Raw, "brace text"),
				})
			}
		}

		return e, nil
	}
}

func And() parser.Func[*ast.And] {
	return func(p *parser.Parser) (*ast.And, *fancyerr.Error) {
		a := &ast.And{And: p.Pos()}
		if !parser.TryToken(p, "&") {
			return nil, &fancyerr.Error{
				Message: "missing &-attributes",
				Primary: quickanno.Expected(p, a.And, "&"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var ok bool
		a.Attributes, ok = parser.TryOk(p, argument.Arguments())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "&-attributes: missing attributes",
				Primary: quickanno.Expected(p, p.Pos(), "attributes"),
				Examples: []fancyerr.Example{
					{Example: "`&(aria-label=\"woof\")`"},
				},
			})
			return a, nil
		}

		return a, nil
	}
}
