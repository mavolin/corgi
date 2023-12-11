package write

import (
	"regexp"
	"strings"

	"github.com/mavolin/corgi/file"
)

var funcHeaderRegexp = regexp.MustCompile(`^func *\w+\([^)]*\) *\{`)

func code(ctx *ctx, c file.Code) {
	var ignoreControl bool
	for _, line := range c.Lines {
		switch {
		// If we're in the body of an inline function, we don't want to flush
		case funcHeaderRegexp.MatchString(line.Code):
			ignoreControl = true
		case strings.Contains(line.Code, "break"),
			strings.Contains(line.Code, "continue"),
			strings.Contains(line.Code, "goto"):
			if !ignoreControl {
				ctx.flushGenerate()
				ctx.flushClasses()
				ctx.callClosedIfClosed()
			}
		}

		ctx.writeln(line.Code)
	}
}
