package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func ConstantString(usedAsPlural string) parser.Func[ast.String] {
	return func(p *parser.Parser) ast.String {
		if s := parser.Try(p, ConstantInterpretedString(usedAsPlural)); s != nil {
			return s
		} else if s := parser.Try(p, ConstantRawString(usedAsPlural)); s != nil {
			return s
		}
		return nil
	}
}

func ConstantInterpretedString(usedAsPlural string) parser.Func[*ast.InterpretedString] {
	return func(p *parser.Parser) *ast.InterpretedString {
		s := parser.Try(p, InterpretedString())
		if s == nil {
			return nil
		}

		for _, n := range s.Contents {
			if _, ok := n.(ast.ConstantInterpretedStringNode); ok {
				continue
			}
			p.CaptureError(nonConstantInConstantStringError(p, n, usedAsPlural))
		}

		return s
	}
}

func ConstantRawString(usedAsPlural string) parser.Func[*ast.RawString] {
	return func(p *parser.Parser) *ast.RawString {
		s := parser.Try(p, RawString())
		if s == nil {
			return nil
		}

		for _, n := range s.Contents {
			if _, ok := n.(ast.ConstantRawStringNode); ok {
				continue
			}
			p.CaptureError(nonConstantInConstantStringError(p, n, usedAsPlural))
		}

		return s
	}
}

func nonConstantInConstantStringError(p *parser.Parser, node ast.Node, usedAsPlural string) *diagnostic.Diagnostic {
	return &diagnostic.Diagnostic{
		Message: "constant string: use of non-constant expression",
		Primary: []diagnostic.Annotation{
			anno.Node(p.File, node, "not a constant"),
		},
		Hints: []diagnostic.Hint{
			{
				Hint: "Strings used as " + usedAsPlural + " must be constant. " +
					"A constant string can only contain text, character escapes, and character references.",
			},
		},
	}
}

func String() parser.Func[ast.String] {
	return func(p *parser.Parser) ast.String {
		switch {
		case parser.MatchesRune(p, '"'):
			return parser.Try(p, InterpretedString())
		case parser.MatchesRune(p, '`'):
			return parser.Try(p, RawString())
		default:
			return nil
		}
	}
}

func InterpretedString() parser.Func[*ast.InterpretedString] {
	return func(p *parser.Parser) *ast.InterpretedString {
		open := parser.TryRuneAt(p, '"')
		if open == nil {
			return nil
		}

		var s ast.InterpretedString
		s.Open = open

		s.Contents = parser.Collect(p, InterpretedStringNode(), nil)

		s.Close = parser.TryRuneAt(p, '"')
		if s.Close == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "string: missing closing quote",
				Primary: quickanno.Expected(p, *s.Open, "a closing quote for the opening quote here"),
			})
		}

		return &s
	}
}

func InterpretedStringNode() parser.Func[ast.InterpretedStringNode] {
	return func(p *parser.Parser) ast.InterpretedStringNode {
		if txt := parser.Try(p, InterpretedStringText()); txt != nil {
			return txt
		} else if inter := parser.Try(p, interpolation.StringInterpolation()); inter != nil {
			return inter
		}
		return nil
	}
}

func InterpretedStringText() parser.Func[*ast.InterpretedStringText] {
	return func(p *parser.Parser) *ast.InterpretedStringText {
		pos := p.Pos()

		text := parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, '#', '\n', '"') ||
				parser.Matches(p, interpolation.UnambiguousHash())
		})
		if text == "" {
			return nil
		}

		return &ast.InterpretedStringText{Text: text, Position: &pos}
	}
}

func RawString() parser.Func[*ast.RawString] {
	return func(p *parser.Parser) *ast.RawString {
		open := parser.TryRuneAt(p, '`')
		if open == nil {
			return nil
		}

		var s ast.RawString
		s.Open = open

		s.Contents = parser.Collect(p, RawStringNode(), nil)

		s.Close = parser.TryRuneAt(p, '`')
		if s.Close == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "string: missing closing quote",
				Primary: quickanno.Expected(p, *s.Open, "a closing quote for the opening quote here"),
			})
		}

		return &s
	}
}

func RawStringNode() parser.Func[ast.RawStringNode] {
	return func(p *parser.Parser) ast.RawStringNode {
		if txt := parser.Try(p, RawStringText()); txt != nil {
			return txt
		} else if inter := parser.Try(p, interpolation.StringInterpolation()); inter != nil {
			return inter
		}
		return nil
	}
}

func RawStringText() parser.Func[*ast.RawStringText] {
	return func(p *parser.Parser) *ast.RawStringText {
		pos := p.Pos()

		var text string
		text = parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, '#', '`') ||
				parser.Matches(p, interpolation.UnambiguousHash())
		})
		if text == "" {
			return nil
		}

		return &ast.RawStringText{Text: text, Position: &pos}
	}
}
