<div align="center">
<h1>corgi</h1>

[![GitBook](https://img.shields.io/badge/Docs-GitBook-blue)](https://mavolin.gitbook.io/corgi)
[![Test](https://github.com/mavolin/corgi/actions/workflows/test.yml/badge.svg)](https://github.com/mavolin/corgi/actions)
[![Code Coverage](https://codecov.io/gh/mavolin/corgi/branch/main/graph/badge.svg?token=ewFEQGgMES)](https://codecov.io/gh/mavolin/corgi)
[![Go Report Card](https://goreportcard.com/badge/github.com/mavolin/corgi)](https://goreportcard.com/report/github.com/mavolin/corgi)
[![License MIT](https://img.shields.io/github/license/mavolin/corgi)](https://github.com/mavolin/corgi/blob/main/LICENSE)
</div>

---

## About

Corgi is an HTML template language, inspired by pug (hence the name), for Go. 
Just like pug, corgi also uses code generation to generate its templates.

## Main Features

* 👀 Highly readable syntax, not just replacing placeholders
* 👪 Support for inheritance
* ➕ Mixins ‒ functions (with parameters) that render repeated pieces of code
* 🗄 Import mixins from other files
* 🖇 Split large templates into multiple files
* ✨ Import any Go package and use its constants, variables, types, and functions—no need for `FuncMap`s or the like
* 🤏 Generates minified HTML (and through the use of filters also minified CSS and JS)

## Getting Started

Have a look at the documentation on [GitBook](https://mavolin.gitbook.io/corgi) to learn corgi.
If you already know pug, you can also find a detailed list of differences there.

## License

Built with ❤️ by [Maximilian von Lindern](https://github.com/mavolin). Available under the [MIT License](./LICENSE).
