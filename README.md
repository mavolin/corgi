<div align="center">
<h1>corgi</h1>

[![GitBook](https://img.shields.io/badge/docs-GitBook-blue)](https://mavolin.gitbook.io/corgi)
[![Test](https://github.com/mavolin/corgi/actions/workflows/test.yml/badge.svg)](https://github.com/mavolin/corgi/actions)
[![Code Coverage](https://codecov.io/gh/mavolin/corgi/branch/develop/graph/badge.svg?token=ewFEQGgMES)](https://codecov.io/gh/mavolin/corgi)
[![Go Report Card](https://goreportcard.com/badge/github.com/mavolin/corgi)](https://goreportcard.com/report/github.com/mavolin/corgi)
[![License MIT](https://img.shields.io/github/license/mavolin/corgi)](./LICENSE)
</div>

---

## About

Corgi is an HTML template language for Go, inspired by pug (hence the name).
Just like pug, corgi also uses code generation to generate its templates.

## Main Features

* 👀 Highly readable syntax that models HTML, not just replacing placeholders
* ➕ Mixins—functions that render repeated pieces of corgi
* 🌀 Conditional classes that are actually readable
* 🗄 Import other files to use their mixins
* 🖇 Split large templates into multiple files
* 👪 Support for inheritance
* ✨ Import any Go package and use any of its types, functions and constants—no need for `FuncMap`s
* 🤏 Generates compile-time minified HTML, CSS, and JS
* 🔒 Context-aware auto-escaping and filtering of HTML, CSS, JS, and special HTML attributes
* 🛡️ Script CSP nonce injection
* ⚠️ Descriptive, Rust-style errors

## Example

First impressions matter, so here is an example of a simple template:

```jade
import "strings"

// corgi has a stdlib with a handful of useful mixins
use "fmt"

// These are the name and params of the function that corgi will generate.
// Besides the params you specify here, corgi will also add an io.Writer that
// it'll write the output to, and an error return, that returns any errors
// it encounters when writing to the io.Writer.
//
// The real signature will look like this:
// func LearnCorgi(w io.Writer, name string, knowsPug bool, friends []string) error
func LearnCorgi(name string, knowsPug bool, friends []string)

mixin greet(name string) Hello, #{name}!

doctype html
html(lang="en")
  head
    title Learn Corgi
  body
    h1 Learn Corgi
    p#greeting.greeting
      // the & allows you to add additional attributes to an element
      if strings.HasPrefix(name, "M"): &.font-size--big
      +greet(name=name)

    p
      if knowsPug
        > #{name}, since you already know pug,
          learning corgi will be even more of #strong[a breeze] for you!#[ ]

      > Head over to #a(href="https://mavolin.gitbook.io/corgi")[GitBook]
        to learn it.

    p And make sure to tell #+fmt.List(val=friends) about corgi too!
```

Pretty-Printed output:

```html
<!-- LearnCorgi(myWriter, "Maxi", true, []string{"Huey", "Dewey", "Louie"}) -->

<!doctype html>
<html lang=en>
<head><title>Learn Corgi</title></head>
<body>
<h1>Learn Corgi</h1>
<p id=greeting class="greeting font-size--big">Hello, Maxi!</p>
<p>
    Maxi, since you already know pug,
    learning corgi will be even more of <strong>a breeze</strong> for you! Head over to
    <a href=https://mavolin.gitbook.io/corgi>GitBook</a> to learn it.
</p>
<p>And make sure to tell Huey, Dewey, and Louie about corgi too!</p>
</body>
</html>
```

> If you're interested in the generated code,  have a look at the `examples` directory.

## Want to know more?

Have a look at the guide on [GitBook](https://mavolin.gitbook.io/corgi).

## License

Built with ❤ by [Maximilian von Lindern](https://github.com/mavolin).
Available under the [MIT License](./LICENSE).
