package fileutil

import (
	"strconv"
	"strings"

	"github.com/mavolin/corgi/file"
)

// Quote returns the [file.String] in quotes, as it was written in the file.
func Quote(s file.String) string {
	return string(s.Quote) + s.Contents + string(s.Quote)
}

// Unquote unquotes the passed [file.String].
// If the passed string is not syntactically valid, Unquote returns an empty
// string.
func Unquote(s file.String) string {
	if s.Quote == '`' {
		return s.Contents
	} else if s.Quote == '"' {
		unq, err := strconv.Unquote(`"` + s.Contents + `"`)
		if err != nil {
			return ""
		}

		return unq
	}

	return ""
}

// UsePath returns the path needed to use this file.
//
// It correctly accounts for files from corgi's standard library.
func UsePath(f *file.File) string {
	if IsStdLibFile(f) {
		if len(f.PathInModule) > len("std/") {
			return f.PathInModule[len("std/"):]
		}

		return ""
	}

	return f.Module + "/" + f.PathInModule
}

// MachineComment represents a corgi comment intended for machines.
// It follows the same semantics as a Go machine comments, namely a comments
// whose text is NOT separated from the `//` by any whitespace.
//
// Each machine comment starts with a namespace and an optional directive,
// separated from the namespace by a colon.
// It is followed by optional args, separated by a space from the
// namespace/directive.
type MachineComment struct {
	Source file.CorgiComment

	Namespace string
	Directive string
	Args      string
}

// ParseMachineComment attempts to parse the passed [file.CorgiComment] as a
// machine comment.
//
// If it is not, ParseMachineComment returns nil.
func ParseMachineComment(c file.CorgiComment) *MachineComment {
	if len(c.Lines) != 1 {
		return nil
	}

	line := c.Lines[0]

	if c.Line != line.Line || c.Col != line.Col-2 {
		return nil
	} else if line.Comment == "" {
		return nil
	}

	namespaceAndDirective, args, _ := strings.Cut(line.Comment, " ")

	var mc MachineComment
	mc.Namespace, mc.Directive, _ = strings.Cut(namespaceAndDirective, ":")
	mc.Args = strings.TrimLeft(args, " ")
	return &mc
}
