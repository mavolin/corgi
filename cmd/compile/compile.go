// Package compile provides the compile subcommand.
package compile

import (
	"flag"

	"github.com/mavolin/corgi/v2/cmd/command"
	"github.com/mavolin/corgi/v2/cmd/command/flags"
)

var (
	meta = command.Meta{
		Name: "compile",
		ArgUsages: [][2]string{
			{"", "read from stdin and write to stdout"},
			{"<file...>", "read files as a single package and write to stdout"},
			{"<dir>", "read package and write to directory"},
			{"<dir>/...", "recursively compile the dir and its subdirs"},
		},
		ShortDescription: "Compile corgi files or directories.",
		LongDescription: "Compiles a list of corgi files or a directory into a single Go file " +
			"(or as few files as possible, in case of conflicting imports).\n" +
			"\n" +
			"In the directory mode, all corgi files (*.corgi) in that directory are " +
			"processed. You may append '/...' to a the directory to recursively compile all " +
			"subdirectories as well.\n" +
			"You may also choose to manually list files instead. While those files may be in " +
			"different directories, they must all share the same package name, as defined by " +
			"the package directive in each file.\n" +
			"\n" +
			"In the directory modes, the output file is placed within that directory, all " +
			"other modes write to stdout. You may change this behavior using the -o and " +
			"-stdout flags.\n" +
			"If the files have conflicting imports, i.e. two files import different packages " +
			"under the same name, compile will create multiple numbered output files. If " +
			"outputting to stdout, compile will fail with an error instead.",
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

	s.StringVar(&f.out, "o", "",
		"The `path` of the output file.\n"+
			"If multiple files need to be generated and the base of the supplied path contains an "+
			"'*', it will be substituted with a number. Existing files matching the pattern are "+
			"deleted before compiling. In modes that output a file the file name defaults to "+
			"'package*.corgi.go'.\n"+
			"You may also specify a directory 'dir' which is equivalent to setting "+
			"'-o dir/package*.corgi.go'.\n"+
			"As a special case in the recursive directory mode (./...), -o accepts a pattern "+
			"instead of a path and uses it to generate the output files, each in their respective "+
			"directories.")
	s.BoolVar(&f.stdout, "stdout", false,
		"Output to stdout, even though not reading from stdin.\n"+
			"Mutually exclusive with -o.")
	s.BoolVar(&f.debug, "debug", false,
		"Whether to include debug information in the output.\n"+
			"Generated code will then contain references to the original corgi\n"+
			"file and line number.")
}

func run(_ *command.Cmd, f *Flags, args []string) {
	return
}
