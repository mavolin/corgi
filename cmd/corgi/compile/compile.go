// Package compile provides the compile subcommand.
package compile

import (
	"flag"

	"github.com/mavolin/corgi/v2/cmd/corgi/command"
	"github.com/mavolin/corgi/v2/cmd/corgi/flags"
)

var (
	meta = command.Meta{
		Name: "compile",
		ArgUsages: []string{
			"             read from stdin and write to stdout",
			"<directory>  read package and write to -o",
			"<file...>    read files as a single package and write to -o",
			"./...        recursively compile pwd and subdirs",
		},
		ShortDescription: "Compile corgi files or directories.",
		LongDescription: `Compiles a list of corgi files or a directory into a single Go file.

In directory mode, all corgi files (*.corgi) in that directory will be compiled
into a single Go file named package.corgi.go and placed in the same directory.
If you set -o to another directory, the file will be placed there instead. If 
-o is set to a file, or the path does not exist, the output will be written to
that file. You may direct the output to stdout using the -stdout flag.

Similarly, if one or more file paths are specified, all corgi files will be
compiled into a single Go file. The files must all be part of the same package,
as defined by their package directive. -o will use the pwd.

In the rare case that two corgi files have conflicting imports, and -o is set
to a directory or not set and we're compiling a directory, the compiler will
create multiple files named package1.corgi.go, etc. Otherwise, i.e. if -o is
set to a file or if writing to stdout, the compiler will stop with an error.

Compile also accepts the special ./... argument, which recursively traverses the
present working directory and generates a Go file for every directory that
contains at least one corgi file. When using ./... -o and -stdout cannot be 
used.`,
	}

	Command = command.Command(meta, new(Flags), run)
)

type Flags struct {
	flags.LoadFlags
	out           string
	skipGoImports bool
	goImportsExec string
	stdout        bool
	deps          bool
	debug         bool
	verbose       bool
	color         bool
	goExec        string
}

func (f *Flags) Bind(s *flag.FlagSet) {
	f.LoadFlags.Bind(s)

	s.StringVar(&f.out, "o", "package.corgi.go",
		"The `path` to the output file or directory.\n"+
			"Defaults to package.corgi.go.")
	s.BoolVar(&f.stdout, "stdout", false,
		"Output to stdout, even though a directory was passed.\n"+
			"If -o is set, this flag is ignored.")
	s.BoolVar(&f.deps, "deps", false,
		"Whether to compile local dependencies of the named files or the directory as well.\n"+
			"Does not work with stdin or the ./... argument.")
	s.BoolVar(&f.debug, "debug", false,
		"Whether to include debug information in the output.\n"+
			"Generated code will then reference the original corgi file and line number.")
	s.BoolVar(&f.skipGoImports, "skip-goimports", false, "Whether to skip goimports.")
	s.StringVar(&f.goImportsExec, "goimports", "goimports",
		"`Path` to the goimports executable.\n"+
			"Defaults to the goimports executable in $PATH.")
}

func run(_ *command.Cmd, f *Flags, args []string) {
	return
}
