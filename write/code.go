package write

import (
	"github.com/mavolin/corgi/file"
)

func code(ctx *ctx, c file.Code) {
	for _, line := range c.Lines {
		ctx.writeln(line.Code)
	}
}
