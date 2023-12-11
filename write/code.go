package write

import (
	"strings"

	"github.com/mavolin/corgi/file"
)

func code(ctx *ctx, c file.Code) {
	for _, line := range c.Lines {
		switch {
		case strings.Contains(line.Code, "break"),
			strings.Contains(line.Code, "continue"),
			strings.Contains(line.Code, "return"),
			strings.Contains(line.Code, "goto"):
			ctx.flushGenerate()
			ctx.flushClasses()
			ctx.callClosedIfClosed()
		}

		ctx.writeln(line.Code)
	}
}
