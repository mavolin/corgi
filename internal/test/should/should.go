package should

import (
	"os"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Equal[T any](t testing.TB, want, got T, opts ...cmp.Option) bool {
	t.Helper()

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		prettyComparison(t, "Equal", "{{Arg1}} != {{Arg2}}", diff)
		return false
	}
	return true
}

func NotEqual[T any](t testing.TB, want, got T, opts ...cmp.Option) bool {
	t.Helper()

	if cmp.Equal(want, got, opts...) {
		prettyComparison(t, "NotEqual", "{{Arg1}} == {{Arg2}}", "")
		return false
	}
	return true
}

func NoError(t testing.TB, err error) bool {
	t.Helper()

	if err != nil {
		prettyComparison(t, "NoError", "{{Arg1}} != nil", err.Error())
		return false
	}

	return true
}

func Error(t testing.TB, want, got error) bool {
	t.Helper()

	if diff := cmp.Diff(want, got, cmpopts.EquateErrors()); diff != "" {
		prettyComparison(t, "Error", "{{Arg1}} != {{Arg2}}", diff)
		return false
	}
	return true
}

func True(t testing.TB, got bool) bool {
	t.Helper()

	if !got {
		prettyCondition(t, "True", true, "")
		return false
	}

	return true
}

func False(t testing.TB, got bool) bool {
	t.Helper()

	if got {
		prettyCondition(t, "False", false, "")
		return false
	}

	return true
}

func prettyComparison(t testing.TB, name, message, extra string) {
	t.Helper()

	path, targetLine, ok := callerInfo(2)
	if !ok {
		t.Error(message)
		return
	}
	lns := lines(path)
	c, cMultiline := comment(lns, targetLine)
	arg1, _, arg2 := args(lns, name, false, targetLine)

	var s strings.Builder
	s.Grow(len(c) + len(message) + len(arg1) + len(arg2) + len(extra))

	writeComment(&s, c, cMultiline)

	i := strings.Index(message, "{{Arg1}}")
	s.WriteString(message[:i])
	s.WriteString(arg1)
	message = message[i+len("{{Arg1}}"):]

	i = strings.Index(message, "{{Arg2}}")
	s.WriteString(message[:i])
	s.WriteString(arg2)
	s.WriteString(message[i+len("{{Arg2}}"):])

	writeExtra(&s, extra)

	t.Error(s.String())
}

func prettyCondition(t testing.TB, name string, want bool, extra string) {
	t.Helper()

	path, targetLine, ok := callerInfo(2)
	if !ok {
		if want {
			t.Error("got == false")
		} else {
			t.Error("got == true")
		}
		return
	}
	lns := lines(path)
	c, cMultiline := comment(lns, targetLine)
	arg1, compOp, arg2 := args(lns, name, true, targetLine)

	var s strings.Builder
	s.Grow(len(c) + len(arg1) + len("== false") + len(extra))

	writeComment(&s, c, cMultiline)

	if compOp != "" {
		s.WriteString(arg1)
		if want {
			switch compOp {
			case "==":
				s.WriteString(" != ")
			case "!=":
				s.WriteString(" == ")
			case "<":
				s.WriteString(" >= ")
			case "<=":
				s.WriteString(" > ")
			case ">":
				s.WriteString(" <= ")
			case ">=":
				s.WriteString(" < ")
			default:
				panic("unknown comparison operator: " + compOp)
			}
		} else {
			s.WriteByte(' ')
			s.WriteString(compOp)
			s.WriteByte(' ')
		}
		s.WriteString(arg2)
	} else {
		if want {
			s.WriteString(arg1)
			s.WriteString(" == false")
		} else {
			s.WriteString(arg1)
			s.WriteString(" == true")
		}
	}

	writeExtra(&s, extra)

	t.Error(s.String())
}

func writeExtra(s *strings.Builder, extra string) {
	if extra != "" {
		s.WriteString(":\n")
		s.WriteString(extra)
	}
}

func writeComment(s *strings.Builder, c string, cMultiline bool) {
	if c != "" {
		if cMultiline {
			s.WriteByte('\n')
			s.WriteString(c)
			s.WriteString(":\n")
		} else {
			s.WriteString(c)
			s.WriteString(": ")
		}
	}
}

func comment(lines []string, target int) (c string, multiline bool) {
	if target < 1 || target > len(lines) {
		return "", false
	}

	// Check for an inline comment
	if _, after, _ := strings.Cut(lines[target-1], "// "); after != "" {
		return after, false
	}

	// Check for comment attached to group
	for _, ln := range slices.Backward(lines[:target]) {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "// ") {
			c = ln[len("// "):] + "\n" + c
			if c != "" {
				multiline = true
			}
		} else if ln == "" {
			break
		}
	}
	if c != "" {
		c = c[:len(c)-1] // remove trailing newline
	}

	return c, multiline
}

func args(lines []string, name string, comparison bool, target int) (arg1, compOp, arg2 string) {
	if target < 1 || target > len(lines) {
		return "", "", ""
	}

	line := strings.TrimSpace(lines[target-1])
	if line == "" {
		return "", "", ""
	}

	startRegexp := regexp.MustCompile(`(?m)[\pL\pN_][ \t]*.\s*` + name + `\([ \t]*(.*)[ \t]*\)`)
	matches := startRegexp.FindStringSubmatch(line)
	if len(matches) != 2 {
		return "", "", ""
	}
	argsStr := matches[1]

	var start int
	var str byte // 0, '\', '"'
	var parenCount int
	i := 0
Args:
	for ; i < len(argsStr); i++ {
		c := argsStr[i]
		switch c {
		case '!', '=', '<', '>':
			isCompOp := c == '<' || c == '>' || (len(argsStr) > i+1 && argsStr[i+1] == '=')
			if comparison && isCompOp {
				arg1 = strings.TrimSpace(argsStr[start:i])
				if len(argsStr) > i+1 && argsStr[i+1] == '=' {
					compOp = string(c) + "="
					start = i + 2
				} else {
					compOp = string(c)
					start = i + 1
				}
				i++
			}
		case ',':
			if str == 0 && parenCount == 0 {
				switch {
				case start == 0:
					start = i + 1
				case arg1 == "":
					arg1 = strings.TrimSpace(argsStr[start:i])
					start = i + 1
				default:
					arg2 = strings.TrimSpace(argsStr[start:i])
					return arg1, compOp, arg2
				}
			}
		case '(':
			if str == 0 {
				parenCount++
			}
		case ')':
			if str == 0 {
				if parenCount == 0 {
					// unmatched closing parenthesis
					break Args
				}
				parenCount--
			}
		case '\\':
			if str != 0 {
				i++ // skip next character
			}
		case '"', '\'':
			switch str {
			case 0:
				str = c
			case c:
				str = 0
			}
		}
	}
	if str == 0 && parenCount == 0 {
		switch {
		case start == 0:
		case arg1 == "":
			arg1 = strings.TrimSpace(argsStr[start:i])
		default:
			arg2 = strings.TrimSpace(argsStr[start:i])
		}
	}

	if arg1 == "" {
		arg1, arg2 = "got", "want"
	}

	return arg1, compOp, arg2
}

func lines(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	return strings.Split(string(data), "\n")
}

func callerInfo(skip int) (path string, line int, ok bool) {
	callers := make([]uintptr, 48)
	n := runtime.Callers(2+skip, callers)

	frames := runtime.CallersFrames(callers[:n])
	var returnNext bool
	for {
		frame, more := frames.Next()
		switch {
		case strings.HasSuffix(frame.File, "test/should/should.go"):
			fallthrough
		case strings.HasSuffix(frame.File, "test/must/must.go"):
			returnNext = true
		case returnNext:
			return frame.File, frame.Line, true
		case !more:
			return "", 0, false
		}
	}
}
