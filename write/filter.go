package write

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/internal/anno"
)

func commandFilter(ctx *ctx, filter file.CommandFilter) {
	if !ctx.allowAllFilters {
		for _, af := range ctx.allowedFilters {
			if af == filter.Name {
				goto allowed
			}
		}

		if ctx.cli {
			fmt.Println((&corgierr.Error{
				Message: "disallowed filter",
				ErrorAnnotation: anno.Anno(ctx.currentFile(), anno.Annotation{
					Start:       filter.Position,
					StartOffset: 1,
					Len:         len(filter.Name),
					Annotation:  "this filter is not allowed to be executed under the current settings",
				}),
				Suggestions: []corgierr.Suggestion{
					{
						Suggestion: "allow this filter using the `-allow-filter` flag\n" +
							"or edit the list of allowed filters using `-edit-allowed-filters`",
					},
				},
			}).Pretty(ctx.corgierrPretty))
		}

		panic(fmt.Errorf("%s:%d:%d: filter `%s` not allowed by settings", ctx.currentFile().Name, filter.Line, filter.Col, filter.Name))
	allowed:
	}

	args := make([]string, len(filter.Args))
	for i, arg := range filter.Args {
		switch arg := arg.(type) {
		case file.RawCommandFilterArg:
			args[i] = arg.Value
		case file.StringCommandFilterArg:
			var err error
			args[i], err = strconv.Unquote(string(arg.Quote) + arg.Contents + string(arg.Quote))
			if err != nil {
				// this should be caught by the parser, so we're fine
				panic(fmt.Errorf("%s:%d:%d: filter arg %d: invalid string", ctx.currentFile().Name, filter.Line, filter.Col, i+1))
			}
		}
	}

	var in bytes.Buffer

	var prevLnNo int
	var n int
	for _, ln := range filter.Body {
		if prevLnNo > 0 {
			n += ln.Position.Line - prevLnNo
		}
		n += len(ln.Line)
	}
	in.Grow(n)
	for _, ln := range filter.Body {
		if prevLnNo > 0 {
			in.WriteString(strings.Repeat("\n", ln.Position.Line-prevLnNo))
		}
		in.WriteString(ln.Line)
	}

	var out bytes.Buffer

	cmd := exec.Command(filter.Name, args...) //nolint:gosec
	cmd.Stdin = &in
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		if ctx.cli {
			fmt.Println((&corgierr.Error{
				Message: "failed to run filter",
				ErrorAnnotation: anno.Anno(ctx.currentFile(), anno.Annotation{
					Start:      filter.Position,
					ToEOL:      true,
					Annotation: err.Error(),
				}),
				Cause: err,
			}).Pretty(ctx.corgierrPretty))
		}

		panic(fmt.Errorf("%s:%d:%d: failed to run filter: %w", ctx.currentFile().Name, filter.Line, filter.Col, err))
	}

	ctx.closeTag()
	ctx.generate(out.String(), nil)
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

	prevLnNo = 0
	for _, ln := range filter.Body {
		if prevLnNo > 0 {
			sb.WriteString(strings.Repeat("\n", ln.Position.Line-prevLnNo))
		}
		sb.WriteString(ln.Line)
		prevLnNo = ln.Position.Line
	}

	ctx.closeTag()
	switch filter.Type {
	case file.RawHTML:
		minified, err := mini.String("text/html", sb.String())
		if err != nil {
			panic(fmt.Errorf("%s:%d:%d: failed to minify in HTML raw filter: %w", ctx.currentFile().Name, filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	case file.RawSVG:
		minified, err := mini.String("image/svg+xml", sb.String())
		if err != nil {
			panic(fmt.Errorf("%s:%d:%d: failed to minify in SVG raw filter: %w", ctx.currentFile().Name, filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	case file.RawJS:
		minified, err := mini.String("application/javascript", sb.String())
		if err != nil {
			panic(fmt.Errorf("%s:%d:%d: failed to minify in JS raw filter: %w", ctx.currentFile().Name, filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	case file.RawCSS:
		minified, err := mini.String("text/css", sb.String())
		if err != nil {
			panic(fmt.Errorf("%s:%d:%d: failed to minify in CSS raw filter: %w", ctx.currentFile().Name, filter.Line, filter.Col, err))
		}

		ctx.generate(minified, nil)
	default:
		ctx.generate(sb.String(), nil)
	}
}
