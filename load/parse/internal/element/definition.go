package element

import (
	"slices"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Definition() parser.Func[*ast.ElementDefinition] {
	return func(p *parser.Parser) (*ast.ElementDefinition, *fancyerr.Error) {
		var def ast.ElementDefinition

		def.Elem = parser.TryKeywordAt(p, "elem", comment.OrAnyWhitespace())
		if def.Elem == nil {
			return nil, &fancyerr.Error{
				Message: "missing element definition",
				Primary: quickanno.Expected(p, p.Pos(), "an element definition"),
			}
		}

		beforePrefix := p.CloneState()
		// technically '(' would be a valid element name, so check that we
		// don't accidentally consume a '(' as a prefix here
		if !parser.MatchesAnyRune(p, '(') {
			def.Prefix = parser.TryOptional(p, Name(), comment.OrAnyWhitespace())
		}

		def.LParen = parser.TryOptionalRuneAt(p, '(', comment.OrAnyWhitespace())
		if def.LParen == nil {
			spec := parser.Try(p, Spec())
			if (spec == nil || len(p.Errors()) > len(beforePrefix.Errors())) && def.Prefix != nil {
				p.RestoreState(beforePrefix)
				def.Prefix = nil
				spec = parser.Must(p, Spec())
			}
			if spec != nil {
				def.Specs = []*ast.ElementSpec{spec}
			}
			return &def, nil
		}

		def.Specs = make([]*ast.ElementSpec, 0, 64)
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
			parser.MustSkip(p, comment.AndEOS())
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
			return nil, &fancyerr.Error{
				Message: "missing closing parenthesis",
				Primary: quickanno.Expected(p, p.Pos(), "a closing parenthesis"),
			}
		}

		return &def, nil
	}
}

func Spec() parser.Func[*ast.ElementSpec] {
	return func(p *parser.Parser) (*ast.ElementSpec, *fancyerr.Error) {
		var s ast.ElementSpec

		s.Name = parser.TryOptional(p, Name(), comment.OrHorizontalWhitespace())
		if s.Name == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "element spec: missing element name",
				Primary: quickanno.Expected(p, p.Pos(), "an element name"),
			})
		}
		s.Type = parser.Try(p, Type())
		if s.Type == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "element spec: missing element type",
				Primary: quickanno.Expected(p, p.Pos(), "an element type"),
			})
		}

		if s.Name == nil && s.Type == nil {
			return nil, &fancyerr.Error{
				Message: "missing element spec",
				Primary: quickanno.Expected(p, p.Pos(), "an element spec"),
				Examples: []fancyerr.Example{
					{Title: "element spec", Example: "div normal"},
				},
			}
		}
		return &s, nil
	}
}

func Type() parser.Func[ast.ElementType] {
	return func(p *parser.Parser) (ast.ElementType, *fancyerr.Error) {
		if bt := parser.Try(p, BasicType()); bt != nil {
			return bt, nil
		} else if at := parser.Try(p, AliasType()); at != nil {
			return at, nil
		}
		return nil, &fancyerr.Error{
			Message: "missing element type",
			Primary: quickanno.Expected(p, p.Pos(), "an element type"),
			Examples: []fancyerr.Example{
				{Title: "named type", Example: "text"},
				{Title: "alias", Example: "= div"},
			},
		}
	}
}

func BasicType() parser.Func[*ast.BasicElementType] {
	return func(p *parser.Parser) (*ast.BasicElementType, *fancyerr.Error) {
		typ := parser.Try(p, TypeName())
		if typ == nil {
			return nil, &fancyerr.Error{
				Message: "missing basic element type",
				Primary: quickanno.Expected(p, p.Pos(), "an element type name"),
			}
		}

		return &ast.BasicElementType{Type: typ}, nil
	}
}

func AliasType() parser.Func[*ast.AliasElementType] {
	return func(p *parser.Parser) (*ast.AliasElementType, *fancyerr.Error) {
		var t ast.AliasElementType

		t.EqualSign = parser.TryRuneAt(p, '=')
		if t.EqualSign == nil {
			return nil, &fancyerr.Error{
				Message: "missing alias element type",
				Primary: quickanno.Expected(p, p.Pos(), "expected an `=` here"),
			}
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		t.Name = parser.Must(p, Name())
		return &t, nil
	}
}

func TypeName() parser.Func[*ast.ElementTypeName] {
	return func(p *parser.Parser) (*ast.ElementTypeName, *fancyerr.Error) {
		var n ast.ElementTypeName
		n.Position = p.PosPtr()

		name := parser.Try(p, golang.Identifier())
		if name == nil {
			return nil, &fancyerr.Error{
				Message: "missing type name",
				Primary: quickanno.Expected(p, *n.Position, "a type name"),
			}
		}
		n.Name = name.Ident
		switch name.Ident {
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
			p.CaptureError(&fancyerr.Error{
				Message: "invalid type name",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, name.Start(), name.End(), "not a valid type name"),
				},
				Hints: []fancyerr.Hint{
					{Hint: "Valid type names are: `void`, `nothing`, `text`, `css`, `js`"},
				},
			})
		}

		return &n, nil
	}
}
