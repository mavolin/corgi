package element

import (
	"slices"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Definition() parser.Func[*ast.ElementDefinition] {
	return func(p *parser.Parser) *ast.ElementDefinition {
		elem := parser.TryKeywordAt(p, "elem", comment.OrAnyWhitespace())
		if elem == nil {
			return nil
		}

		var def ast.ElementDefinition
		def.Elem = elem

		beforePrefix := p.CloneState()
		// technically '(' would be a valid element name, so check that we
		// don't accidentally consume a '(' as a prefix here
		if !parser.MatchesAnyRune(p, '(') {
			def.Prefix = parser.TryOptional(p, Name(), comment.OrAnyWhitespace())
		}

		def.LParen = parser.TryOptionalRuneAt(p, '(', comment.OrAnyWhitespace())
		if def.LParen == nil {
			spec := parser.Try(p, Spec())
			if (spec == nil || p.NumErrors() > beforePrefix.NumErrors()) && def.Prefix != nil {
				p.RestoreState(beforePrefix)
				def.Prefix = nil
				spec = parser.Try(p, Spec())
				if spec == nil {
					p.CaptureError(&diagnostic.Diagnostic{
						Message:  "element definition: missing element name",
						Primary:  quickanno.Expected(p, p.Pos(), "an element name"),
						Examples: []diagnostic.Example{{Example: "`div normal`"}},
					})
				}
			}
			if spec != nil {
				def.Specs = []*ast.ElementSpec{spec}
			}
			return &def
		}

		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			spec := parser.Try(p, Spec())
			if spec == nil {
				break
			}
			def.Specs = append(def.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}
			parser.Try(p, comment.AndMustEOS())
		}
		if len(def.Specs) == 0 {
			def.Specs = nil
		} else {
			def.Specs = slices.Clip(def.Specs)
		}

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "element definition: unexpected runes"
			p.CaptureError(err)
		}

		def.RParen = parser.TryOptionalRuneAt(p, ')', nil)
		if def.RParen == nil {
			return nil
		}

		return &def
	}
}

func Spec() parser.Func[*ast.ElementSpec] {
	return func(p *parser.Parser) *ast.ElementSpec {
		name := parser.TryOptional(p, Name(), comment.OrHorizontalWhitespace())
		if name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "element spec: missing element name",
				Primary: quickanno.Expected(p, p.Pos(), "an element name"),
			})
		}

		typ := parser.Try(p, Type())
		if typ == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "element spec: missing element type",
				Primary: quickanno.Expected(p, p.Pos(), "an element type"),
			})
		}

		if name == nil && typ == nil {
			return nil
		}

		return &ast.ElementSpec{
			Name: name,
			Type: typ,
		}
	}
}

func Type() parser.Func[ast.ElementType] {
	return func(p *parser.Parser) ast.ElementType {
		if bt := parser.Try(p, BasicType()); bt != nil {
			return bt
		} else if at := parser.Try(p, AliasType()); at != nil {
			return at
		}
		return nil
	}
}

func BasicType() parser.Func[*ast.BasicElementType] {
	return func(p *parser.Parser) *ast.BasicElementType {
		typ := parser.Try(p, TypeName())
		if typ == nil {
			return nil
		}

		return &ast.BasicElementType{Type: typ}
	}
}

func AliasType() parser.Func[*ast.AliasElementType] {
	return func(p *parser.Parser) *ast.AliasElementType {
		equalSign := parser.TryRuneAt(p, '=')
		if equalSign == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		var t ast.AliasElementType
		t.EqualSign = equalSign

		t.Name = parser.Try(p, Reference())
		if t.Name == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "alias type: missing element type",
				Primary: quickanno.Expected(p, p.Pos(), "an element reference"),
				Examples: []diagnostic.Example{
					{Example: "`= div`"},
					{Example: "`= mypkg.div`"},
				},
			})
		}

		return &t
	}
}

func TypeName() parser.Func[*ast.ElementTypeName] {
	return func(p *parser.Parser) *ast.ElementTypeName {
		pos := p.Pos()

		name := parser.Try(p, golang.Identifier())
		if name == nil {
			return nil
		}

		var n ast.ElementTypeName
		n.Position = &pos

		n.Name = name.Name
		switch name.Name {
		case "void":
			n.Type = elemtype.Void
		case "nothing":
			n.Type = elemtype.Nothing
		case "normal":
			n.Type = elemtype.Normal
		case "text":
			n.Type = elemtype.Text
		case "css":
			n.Type = elemtype.CSS
		case "js":
			n.Type = elemtype.JS
		default:
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "invalid type name",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, name.Start(), name.End(), "not a valid type name"),
				},
				Hints: []diagnostic.Hint{
					{Hint: "Valid type names are: `void`, `nothing`, `normal`, `text`, `css`, `js`"},
				},
			})
		}

		return &n
	}
}
