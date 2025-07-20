package attribute

import (
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Type() parser.Func[*ast.AttributeType] {
	return func(p *parser.Parser) (*ast.AttributeType, *diagnostic.Diagnostic) {
		var t ast.AttributeType

		t.Quote = parser.TryRuneAt(p, '\'')
		if t.Quote == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute type",
				Primary: quickanno.Expected(p, p.Pos(), "a single quote and then an attribute type"),
			}
		}

		t.Name = parser.Try(p, TypeName())
		if t.Name == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute type name",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute type name"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		// make sure this isn't actually a rune literal
		if parser.MatchesToken(p, "'") {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute type",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, *t.Quote, p.Pos(),
						"expected a single quote and then an attribute type, but found a rune literal instead"),
				},
			}
		}

		t.LBracket = parser.TryRuneAt(p, '[')
		if t.LBracket == nil {
			return &t, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		t.Attribute = parser.Must(p, Name())

		parser.TrySkip(p, comment.OrAnyWhitespace())
		t.RBracket = parser.TryRuneAt(p, ']')
		if t.RBracket == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "missing closing bracket",
				Primary: quickanno.Expected(p, p.Pos(), "a closing bracket"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *t.LBracket, "because of this opening bracket"),
				},
			})
			return &t, nil
		}

		return &t, nil
	}
}

func TypeName() parser.Func[*ast.AttributeTypeName] {
	return func(p *parser.Parser) (*ast.AttributeTypeName, *diagnostic.Diagnostic) {
		var n ast.AttributeTypeName
		n.Position = p.PosPtr()

		ident := parser.Try(p, golang.Identifier())
		if ident == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing attribute type name",
				Primary: quickanno.Expected(p, *n.Position, "an attribute type name"),
			}
		}
		n.Name = ident.Name

		switch n.Name {
		case "unsafe":
			n.Type = attrtype.Unsafe
		case "unsafeBool":
			n.Type = attrtype.UnsafeBool
		case "bool":
			n.Type = attrtype.Bool
		case "text":
			n.Type = attrtype.Text
		case "innocuous":
			n.Type = attrtype.Innocuous
		case "css":
			n.Type = attrtype.CSS
		case "js":
			n.Type = attrtype.JS
		case "url":
			n.Type = attrtype.URL
		case "urlList":
			n.Type = attrtype.URLList
		case "resourceURL":
			n.Type = attrtype.ResourceURL
		case "srcset":
			n.Type = attrtype.Srcset
		default:
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "unknown attribute type",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, ident.Start(), ident.End(), "not a known attribute type"),
				},
				Hints: []diagnostic.Hint{
					{
						Hint: "Valid attribute types are: " +
							"`unsafe`, `unsafeBool`, `bool`, `text`, `innocuous`," +
							" `css`, `js`, `url`, `urlList`, `resourceURL`, `srcset`",
					},
				},
			})
		}
		return &n, nil
	}
}
