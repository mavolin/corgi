package write

import (
	"path"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
)

func include(ctx *ctx, incl file.Include) {
	ctx.closeStartTag()

	switch inclF := incl.Include.(type) {
	case file.CorgiInclude:
		scope(ctx, inclF.File.Scope, true)
	case file.OtherInclude:
		var contents string

		switch path.Ext(fileutil.Unquote(incl.Path)) {
		case ".js":
			var err error
			contents, err = mini.String("application/javascript", inclF.Contents)
			if err != nil {
				contents = inclF.Contents
			}
		case ".css":
			var err error
			contents, err = mini.String("text/css", inclF.Contents)
			if err != nil {
				contents = inclF.Contents
			}
		case ".html":
			var err error
			contents, err = mini.String("text/html", inclF.Contents)
			if err != nil {
				contents = inclF.Contents
			}
		default:
			contents = plainBodyTextEscaper.f(inclF.Contents)
		}

		ctx.generate(contents, nil)
	}
}
