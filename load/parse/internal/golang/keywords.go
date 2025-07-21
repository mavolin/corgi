package golang

import (
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// IsKeyword returns whether the given string is a Go keyword.
func IsKeyword(s string) bool { // https://go.dev/ref/spec#Keywords
	switch s {
	case "break":
	case "case":
	case "chan":
	case "const":
	case "continue":
	case "default":
	case "defer":
	case "else":
	case "fallthrough":
	case "for":
	case "func":
	case "go":
	case "goto":
	case "if":
	case "import":
	case "interface":
	case "map":
	case "package":
	case "range":
	case "return":
	case "select":
	case "struct":
	case "switch":
	case "type":
	case "var":
	// corgi
	case "comp":
	case "with":
	case "state":
	case "attr":
	case "elem":
	case "ordered":
	default:
		return false
	}
	return true
}

func Keyword() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		switch {
		case parser.TryToken(p, "break") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "break", nil
		case parser.TryToken(p, "case") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "case", nil
		case parser.TryToken(p, "chan") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "chan", nil
		case parser.TryToken(p, "const") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "const", nil
		case parser.TryToken(p, "continue") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "continue", nil
		case parser.TryToken(p, "default") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "default", nil
		case parser.TryToken(p, "defer") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "defer", nil
		case parser.TryToken(p, "else") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "else", nil
		case parser.TryToken(p, "fallthrough") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "fallthrough", nil
		case parser.TryToken(p, "for") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "for", nil
		case parser.TryToken(p, "func") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "func", nil
		case parser.TryToken(p, "go") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "go", nil
		case parser.TryToken(p, "goto") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "goto", nil
		case parser.TryToken(p, "if") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "if", nil
		case parser.TryToken(p, "import") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "import", nil
		case parser.TryToken(p, "interface") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "interface", nil
		case parser.TryToken(p, "map") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "map", nil
		case parser.TryToken(p, "package") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "package", nil
		case parser.TryToken(p, "range") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "range", nil
		case parser.TryToken(p, "return") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "return", nil
		case parser.TryToken(p, "select") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "select", nil
		case parser.TryToken(p, "struct") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "struct", nil
		case parser.TryToken(p, "switch") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "switch", nil
		case parser.TryToken(p, "type") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "type", nil
		case parser.TryToken(p, "var") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "var", nil
		default:
			return "", &diagnostic.Diagnostic{
				Message: "missing keyword",
				Primary: quickanno.Expected(p, p.Pos(), "a keyword"),
			}
		}
	}
}
