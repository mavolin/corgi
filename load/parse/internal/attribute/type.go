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

var basicTypes = func() map[string]attrtype.Type {
	m := make(map[string]attrtype.Type)
	for _, t := range attrtype.All {
		switch t.(type) {
		case attrtype.SpaceList:
		case attrtype.CommaList:
		default:
			m[t.String()] = t
		}
	}
	return m
}()

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

		if t := basicTypes[n.Name]; t != nil {
			n.Type = t
			return &n
		}

		switch n.Name {
		case "spaceList":
			elem := parser.Try(p, listTypeBrackets[attrtype.SpaceListElement](&n, "space list"))
			n.Type = attrtype.SpaceList{Element: elem}
		case "commaList":
			elem := parser.Try(p, listTypeBrackets[attrtype.CommaListElement](&n, "comma list"))
			n.Type = attrtype.CommaList{Element: elem}
		default:
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "unknown attribute type",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, ident.Start(), ident.End(), "not a known attribute type"),
				},
				Hints: []diagnostic.Hint{
					{
						Hint: "Valid attribute types are: " +
							"`unsafe`, `unsafeBool`, `bool`, `text`, `string`," +
							" `css`, `js`, `url`, `urlList`, `resourceURL`, `srcset`",
					},
				},
			})
		}
		return &n
	}
}

func listTypeBrackets[E comparable](n *ast.AttributeTypeName, name string) parser.Func[E] {
	return func(p *parser.Parser) E {
		var zero E
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		n.LBracket = parser.TryRuneAt(p, '[')
		if n.LBracket == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "attribute type: " + name + ": missing opening bracket",
				Primary:  quickanno.Expected(p, p.Pos(), "an opening bracket for the element type"),
				Examples: []diagnostic.Example{{Example: "`" + name + "[string]`"}},
			})
			return zero
		}

		var e E
		parser.TrySkip(p, comment.OrAnyWhitespace())
		elemName := parser.Try(p, golang.Identifier())
		if elemName == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute type: " + name + ": missing element type",
				Primary: quickanno.Expected(p, p.Pos(), "an element type"),
			})
		} else {
			n.Element = elemName.Name
			t := basicTypes[n.Element]
			if t == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "attribute type: " + name + ": unknown element type",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, elemName.Start(), elemName.End(), "not a known attribute type"),
					},
					Examples: []diagnostic.Example{{Example: "`spaceList[string]`"}},
				})
			} else if e, _ = t.(E); e == zero {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "attribute type: spaceList: invalid element type",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, elemName.Start(), elemName.End(), "not a valid "+name+" element type"),
					},
					Explanation: "Not all attribute types can be used as elements in a " + name + ".",
				})
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		n.RBracket = parser.TryRuneAt(p, ']')
		if n.RBracket == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "attribute type: " + name + ": missing closing bracket",
				Primary: quickanno.Expected(p, p.Pos(), "a closing bracket for the element type"),
			})
		}

		return e
	}
}
