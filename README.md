<div align="center">
<h1>corgi</h1>

[![GitBook](https://img.shields.io/badge/Docs-GitBook-blue)](https://mavolin.gitbook.io/corgi)
[![Test](https://github.com/mavolin/corgi/actions/workflows/test.yml/badge.svg)](https://github.com/mavolin/corgi/actions)
[![Code Coverage](https://codecov.io/gh/mavolin/corgi/branch/develop/graph/badge.svg?token=ewFEQGgMES)](https://codecov.io/gh/mavolin/corgi)
[![Go Report Card](https://goreportcard.com/badge/github.com/mavolin/corgi)](https://goreportcard.com/report/github.com/mavolin/corgi)
[![License MIT](https://img.shields.io/github/license/mavolin/corgi)](https://github.com/mavolin/corgi/blob/develop/LICENSE)
</div>

---

## About

Corgi is an HTML template language for Go, inspired by pug (hence the name).
Just like pug, corgi also uses code generation to generate its templates.

## Main Features

* 👀 Highly readable syntax, not just replacing placeholders
* 👪 Support for inheritance
* ➕ Mixins—functions that render repeated pieces of corgi
* 🗄 Import mixins from other files, just like you would import Go packages
* 🖇 Split large templates into multiple files
* ✨ Import any Go package and use its constants, variables, types, and functions—no need for `FuncMap`s or the like
* 🤏 Generates minified HTML (and through the use of filters also minified CSS and JS)
* 🔒 Automatically escapes interpolated HTML, CSS and JS

## Example

First impressions matter, so here is an example of a simple template:

```jade
import "strings"

//- These are the name and params of the function that corgi will generate.
//- Besides the params you specify here, corgi will also add an io.Writer that
//- it'll write the output to, and an error return, that returns any errors
//- it encounters when writing to the io.Writer.
//-
//- The real signature will look like this:
//- func LearnCorgi(__w io.Writer, name string, knowsPug bool, friends []string) error
func LearnCorgi(name string, knowsPug bool, friends []string)

mixin greet(name string)
  | Hello, #{name}!

html(lang="en")
  head
    title Learn Corgi
  body
    h1 Learn Corgi
    p#features.font-size--big
      +greet(name=name)

    p
      if knowsPug
        | #{name}, since you already know pug,
        | learning corgi will be even more of #strong[a breeze] for you!
        |

      
      | Head over to
      | #a(href="https://mavolin.gitbook.io/corgi")[GitBook]
      | to learn it.

    switch len(friends)
      case 0
      case 1
        p And make sure to tell #{friends[0]} about corgi too!
      case 2
        p And make sure to tell #{friends[0]} and #{friends[1]} about corgi too!
      default
        p.
          And make sure to tell
          #{strings.Join(friends[:len(friends)-1], ", ")},
          and #{friends[len(friends)-1]} about corgi too!
```

Pretty-Printed output:

```html
<!-- LearnCorgi(myWriter, "Maxi", true, []string{"Huey", "Dewey", "Louie"}) -->

<!doctype html>
<html lang="en">
<head>
    <title>Learn Corgi</title>
</head>
<body>
<h1>Learn Corgi</h1>
<p id="features" class="font-size--big">Hello, Maxi!</p>
<p>
    Maxi, since you already know pug,
    learning corgi will be even more of <strong>a breeze</strong> for you!
    Head over to <a href="https://mavolin.gitbook.io/corgi">GitBook</a> to
    learn it.
</p>
<p>And make sure to tell Huey, Dewey, and Louie about corgi too!</p>
</body>
</html>
```

## Want to know more?

Have a look at the guide on [GitBook](https://mavolin.gitbook.io/corgi).
If you already know pug, you can also find a detailed list of differences there.

## License

Built with ❤ by [Maximilian von Lindern](https://github.com/mavolin).
Available under the [MIT License](./LICENSE).
