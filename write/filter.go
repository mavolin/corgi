package write

import (
	"fmt"
	"strings"

	"github.com/mavolin/corgi/file"
)

func commandFilter(ctx *ctx, filter file.CommandFilter) {
	panic("implement me")
}

func rawFilter(ctx *ctx, filter file.RawFilter) {
	var n int

	var prevLnNo int
	for _, ln := range filter.Body {
		if prevLnNo > 0 {
			n += ln.Position.Line - prevLnNo
		}
		n += len(ln.Line)
		prevLnNo = ln.Position.Line
	}

	var sb strings.Builder
	sb.Grow(n)

	for _, ln := range filter.Body {
		if prevLnNo > 0 {
			sb.WriteString(strings.Repeat("\n", ln.Position.Line-prevLnNo))
		}
		sb.WriteString(ln.Line)
		prevLnNo = ln.Position.Line
	}

	switch filter.Type {
	case file.RawHTML:
		minified, err := mini.String("text/html", sb.String())
		if err != nil {
			panic(fmt.Errorf("%d:%d: failed to minify in HTML raw filter: %w", filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	case file.RawSVG:
		minified, err := mini.String("image/svg+xml", sb.String())
		if err != nil {
			panic(fmt.Errorf("%d:%d: failed to minify in SVG raw filter: %w", filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	case file.RawJS:
		minified, err := mini.String("application/javascript", sb.String())
		if err != nil {
			panic(fmt.Errorf("%d:%d: failed to minify in JS raw filter: %w", filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	case file.RawCSS:
		minified, err := mini.String("text/css", sb.String())
		if err != nil {
			panic(fmt.Errorf("%d:%d: failed to minify in CSS raw filter: %w", filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	default:
		ctx.generate(sb.String(), nil)
	}
}
