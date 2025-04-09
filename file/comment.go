package file

import (
	"strings"

	"github.com/mavolin/corgi/v2/file/ast"
)

// IsCommentDirective returns true, if c is a comment directive.
//
// See [ParseCommentDirective] for more information.
func IsCommentDirective(c *ast.Comment) bool {
	if c == nil || c.Comment == "" {
		return false
	}
	f := c.Comment[0]
	return f != ' ' && f != '\t' && f != '\r' && f != '\n' && f != '*'
}

type CommentDirective struct {
	Namespace string
	Directive string
	Inherited bool
	Args      string
	Comment   string
}

// ParseCommentDirective parses a comment directive, i.e. a comment intended
// for tooling.
//
// Every comment whose first character matches the character class [^ \t\r\n*]
// is considered a comment directive.
//
// It consists of an (optional, though recommended) namespace, a directive, and
// a (usually space-separated) list of arguments.
// If the directive is followed by a '*', it is inherited by the nodes in the
// annotated node's body.
// A comment directive may contain a trailing human-readable comment,
// preceded by " // ".
//
// Examples for comment directives are:
//
//	//namespace:directive arg1 arg2
//	//directive arg1 arg2
//	//namespace:directive
//	//directive
//	//directive // this does something
//	//namespace:directive* foo
func ParseCommentDirective(c *ast.Comment) *CommentDirective {
	if !IsCommentDirective(c) {
		return nil
	}

	namespaceAndDirective, argsAndComment, _ := strings.Cut(c.Comment, " ")

	var ok bool
	namespace, directive, ok := strings.Cut(namespaceAndDirective, ":")
	if !ok {
		namespace, directive = "", namespace
	}

	inherited := strings.HasSuffix(directive, "*")
	if inherited {
		directive = directive[:len(directive)-len("*")]
	}

	args, comment, _ := strings.Cut(argsAndComment, " // ")

	return &CommentDirective{
		Namespace: namespace,
		Directive: directive,
		Inherited: inherited,
		Args:      args,
		Comment:   comment,
	}
}
