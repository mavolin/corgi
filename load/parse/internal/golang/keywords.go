package golang

import (
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
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
	case "block":
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
	return func(p *parser.Parser) string {
		switch {
		case parser.TryToken(p, "break") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "break"
		case parser.TryToken(p, "case") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "case"
		case parser.TryToken(p, "chan") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "chan"
		case parser.TryToken(p, "const") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "const"
		case parser.TryToken(p, "continue") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "continue"
		case parser.TryToken(p, "default") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "default"
		case parser.TryToken(p, "defer") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "defer"
		case parser.TryToken(p, "else") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "else"
		case parser.TryToken(p, "fallthrough") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "fallthrough"
		case parser.TryToken(p, "for") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "for"
		case parser.TryToken(p, "func") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "func"
		case parser.TryToken(p, "go") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "go"
		case parser.TryToken(p, "goto") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "goto"
		case parser.TryToken(p, "if") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "if"
		case parser.TryToken(p, "import") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "import"
		case parser.TryToken(p, "interface") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "interface"
		case parser.TryToken(p, "map") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "map"
		case parser.TryToken(p, "package") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "package"
		case parser.TryToken(p, "range") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "range"
		case parser.TryToken(p, "return") && parser.MatchesWS(p, comment.OrHorizontalWhitespace()):
			return "return"
		case parser.TryToken(p, "select") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "select"
		case parser.TryToken(p, "struct") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "struct"
		case parser.TryToken(p, "switch") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "switch"
		case parser.TryToken(p, "type") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "type"
		case parser.TryToken(p, "var") && parser.MatchesWS(p, comment.OrAnyWhitespace()):
			return "var"
		default:
			return ""
		}
	}
}
