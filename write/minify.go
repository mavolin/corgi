package write

import (
	"regexp"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/svg"
)

var mini = minify.New()

func init() {
	mini.AddFunc("text/css", css.Minify)
	mini.AddFunc("text/html", html.Minify)
	mini.AddFunc("image/svg+xml", svg.Minify)
	mini.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
}
