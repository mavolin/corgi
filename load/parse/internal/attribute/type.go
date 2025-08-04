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
	return func(p *parser.Parser) *ast.AttributeType {
		quote := parser.TryRuneAt(p, '\'')
		if quote == nil {
			return nil
		}

		name := parser.Try(p, TypeName())
		if name == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		// make sure this isn't actually a rune literal
		if parser.MatchesToken(p, "'") {
			return nil
		}

		var t ast.AttributeType
		t.Quote = quote
		t.Name = name

		t.LBracket = parser.TryRuneAt(p, '[')
		if t.LBracket == nil {
			return &t
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		pos := p.Pos()
		t.Attribute = parser.Try(p, Name())
		if t.Attribute == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute type: missing attribute name",
				Primary: quickanno.Expected(p, pos, "an attribute name"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *t.LBracket, "because of this opening bracket"),
				},
			})
			return &t
		}

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
			return &t
		}

		return &t
	}
}

func TypeName() parser.Func[*ast.AttributeTypeName] {
	return func(p *parser.Parser) *ast.AttributeTypeName {
		pos := p.Pos()

		ident := parser.Try(p, golang.Identifier())
		if ident == nil {
			return nil
		}

		var n ast.AttributeTypeName
		n.Position = &pos
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
		return &n
	}
}
