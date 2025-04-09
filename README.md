<div align="center">
<h1>corgi</h1>

[![GitBook](https://img.shields.io/badge/docs-GitBook-blue)](https://mavolin.gitbook.io/corgi)
[![Tests](https://github.com/mavolin/corgi/actions/workflows/test.yml/badge.svg)](https://github.com/mavolin/corgi/actions)
[![Code Coverage](https://codecov.io/gh/mavolin/corgi/branch/v2/graph/badge.svg?token=ewFEQGgMES)](https://codecov.io/gh/mavolin/corgi)
[![Go Report Card](https://goreportcard.com/badge/github.com/mavolin/corgi)](https://goreportcard.com/report/github.com/mavolin/corgi/v2)
[![License MIT](https://img.shields.io/github/license/mavolin/corgi)](./LICENSE)
</div>

---

## About

Corgi is a component-driven HTML template language made for Go.
Corgi puts strong emphasis on readability, security, and ease of use.

## Main Features

* 💪 Powerful and intuitive: Work with it, not around it
* 👀 Highly readable syntax that models HTML, leading to many features not possible in other template languages
* 🌀 Conditional classes that are actually readable
* ➕ Looks like a Go file, with function-like components at the center
* 🔒️ Strong security model, with content-type-aware¹ auto-escaping and filtering—No heuristics, no surprises
* 👍️ `if`, `for` and `switch` are written and work exactly as in Go; no need to learn the quirks of a custom syntax
* ⏱️ Naturally benefits from Go's type safety, due to code generation
* 🛡️ Automatic CSP nonce injection
* ✨ Import any Go package and use any of its types, functions and constants—no need for `FuncMap`s
* ➕ Easy optional chaining, zero coalescing, and string interpolation in Go expressions
* 🤏 Generates compile-time minified HTML, CSS, and JS²
* 🧹 Built-in formatter (WIP)
* ⚠️ Descriptive, Rust-style errors
* ✨ Ideal for HTMX

<small>
¹: Expressions in attributes are escaped through element- and attribute-name-based classification into 
unsafe, bool, text, CSS, JS, URL, resource URL, URL list, and srcset attributes through an allowlist that
covers all HTML5 attributes and all HTMX attributes (including core extensions) by default.
You can trivially extend this allowlist to cover custom attributes.
Unknown attributes, such as data-* attributes, must be explicitly assigned a type if they have a non-constant value.
This is checked at compile time.
<br>
Likewise, the escape strategy for the content of an element is determined by the element's name.
Elements are classified in an allowlist into void, nothing (`iframe`), normal, text, CSS, and JS elements.
Like with attributes, you may extend this list, if you require.
Unknown elements lead to a compile-time error and must be explicitly assigned a type.
<br>
²: HTML is always minified. 
Currently, only JS that contains only text and interpolation can be minified, which should cover most cases.
</small>

## Example

First impressions matter, so here is an example of a simple template:

```go
package example

import (
    "strings"

    // corgi has a stdlib with a couple of useful components
    "corgi/fmt"
)

// A component is the basic building block of corgi, akin to a function.
// In fact, this and all other components get compiled to functions.
comp HelloWorld(name string) {
    !doctype(html)
    html(lang="en") { // <html lang="en">
        head { // <head>
            title [ Hello World! ]
        }
        body {
            // <p class="hello-world">
            p(.hello-world) [ Hello World! ]
            // <p id="greeting">
            p(#greeting) [ How are you, #{name}? ]
        }
    } // </html>
}

// Components can also be passed blocks, essentially HTML placeholders, to use
// in their body.
// base accepts two blocks, title and body.
//
// This makes it easy to create components used as layouts.
// It also replaces the inheritance model some template languages have: 
// Everything is simply a component.
comp base() {
    !doctype(html)
    html(lang="en") {
        head {
            link(rel="stylesheet", href="style.css")
            title { block title }
        }
          
        body { 
            h1 { block title }
            block body
        }
    }
}

// Components have a couple of neat features, like default parameters.
// In the example below, friends has the default value nil, and needn't be 
// passed when calling the component.
//
// To save ourselves a top-level call to base, which adds an unnecessary level 
// of indentation, we can place the call to base directly behind LearnCorgi's
// signature.
comp LearnCorgi(name string, friends []string: nil) : base {
    with title [ Learn Corgi ]
    with body {
        p(.greeting) { // <p class="greeting"> 
            :bigM(name: name) // can modify the parent element: <p class="greeting font-size--big">
            :greeting(name: name)
        }
        
        p {
            if len(friends) >= 3 [
                You have a lot of friends, #{name}!
                Make sure to tell #:fmt.List(val: friends) about corgi too!
            ]
            
            // another way to write text, besides [ bracket text ], are arrow blocks
            > Head over to #a(href="https://mavolin.gitbook.io/corgi")[GitBook]
              to learn it.
        }
    }
}

// bigM conditionally adds the font-size--big class to the calling element,
// if name starts with the letter 'M'.
comp bigM(name string) {
    if strings.HasPrefix(name, "M") {
        &(.font-size--big)
    }
}

comp greeting(name string) [ Hello, #{name}! ]
```

Pretty-Printed output:

```html
<!-- HelloWorld(name: "Vinnie") -->
<!doctype html>
<html lang="en">
<head>
  <title>Hello World</title>
</head>
<body>
  <p class="hello-world">Hello World!</p>
  <p id="greeting">How are you, Vinnie?</p>
</body>
</html>

<!-- LearnCorgi(name: "Maxi", friends: []string{"Huey", "Dewey", "Louie"}) -->

<!doctype html>
<html lang=en>
<head>
  <link rel=stylesheet href=style.css>
  <title>Learn Corgi</title>
</head>
<body>
  <h1>Learn Corgi</h1>
  <p id=greeting class="greeting font-size--big">
    Hello, Maxi!
  </p>
  <p>
    You have a lot of friends, Maxi!
    Make sure to tell Huey, Dewey, and Louie about corgi too!.
    Head over to <a href=https://mavolin.gitbook.io/corgi>GitBook</a> to learn it.
  </p>
</body>
</html>
```

> If you're interested in the generated code, have a look at the `examples` directory.

## Why a Meta-Language?

You might look at the example and wonder why corgi doesn't just build upon good ol' HTML.

Corgi's special syntax brings with it a couple of advantages, that imo outweigh the,
at first glance, unfamiliar syntax:

1. **Mental Concept**: Traditional template languages, using HTML syntax, are usually just text and placeholders.
    They don't manipulate the HTML you write and just interpolate values.
    Components, on the other hand, can manipulate their parents or can get passed attributes that are added to its own
    elements.
    This, obviously, stands in opposition of the traditional placeholder filling.
    A different syntax loosens the expectations one would have when seeing HTML and makes it easier to understand what's
    going on.
    Many of corgi's syntax sugars would also look out of place in HTML. 
    Just imagine the above `bigM` example in Go's template language: It wouldn't feel natural.
2. **Context switching**: Corgi intentionally distinguishes between the `[]` and `{}` brackets for text and code.
    This alone means no escapes for control structures, leading to a better signal-to-noise ratio and reducing the
    cognitive complexity put on the developer.
    Using the HTML syntax would mean giving up that advantage and bringing back many of the pains of other template 
    languages.
3. **Readability**: Also, let's be real for a sec: HTML is really bloated.
    It has high redundancy, is arguably not that readable, and the added syntax of most template languages doesn't help.
    Corgi has less redundancy and a more concise syntax, that needs fewer escapes and keyboard acrobatics for 
    simple features.
4. **Simplicity**: HTML is a pretty simple language, so naturally it is extremely easy to transfer you knowledge of
    HTML to corgi, since syntax-wise there is not that much to learn.
    Once you know how to write elements and attributes, you're already good to go.

## Want to Know More?

Have a look at the [documentation](https://corgi.mavolin.co).

## License

Built with ❤ by [Maximilian von Lindern](https://github.com/mavolin).
Available under the [MIT License](./LICENSE).
