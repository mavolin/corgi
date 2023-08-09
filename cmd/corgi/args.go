package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mavolin/corgi"
	"github.com/mavolin/corgi/internal/meta"
)

var (
	// Misc

	IsGoGenerate bool

	// Flags

	Package string

	PrecompileLibrary bool
	NoGoImports       bool

	OutFile   string
	UseStdout bool

	ScriptNonce bool

	GoExecPath string

	Verbose bool
	Debug   bool

	ForceColorSetting bool
	Color             bool

	TrustedExecutables  []string
	TrustAllExecutables bool

	// Args

	InFile string
	InData []byte
)

func init() {
	IsGoGenerate = os.Getenv("GOFILE") != ""

	var (
		showHelp    bool
		showVersion bool
	)

	flag.Usage = usage

	flag.BoolVar(&showHelp, "h", false, "show this help message")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.StringVar(&Package, "package", "",
		"the name of the package to generate into (default: GOPACKAGE, or pwd)\n"+
			"ignored if -lib is set")
	flag.StringVar(&OutFile, "o", "", "write output to `File` instead of stdout (defaults to `FILE.go`)")
	flag.BoolVar(&UseStdout, "stdout", false, "write to stdout instead of a file")

	flag.BoolVar(&PrecompileLibrary, "lib", false,
		"treat the input file as a library dir, not compatible with stdin;\n"+
			"`-o`, if not set, will default to `"+corgi.PrecompFileName+"`")
	flag.BoolVar(&NoGoImports, "nogoimports", false, "do not run goimports on the generated file")
	flag.StringVar(&GoExecPath, "go", "", "path to the go executable, defaults to a PATH lookup")
	flag.BoolVar(&Verbose, "v", false, "enable verbose output to stderr")
	flag.BoolVar(&Debug, "debug", false, "print file and line information as comments in the generated function")
	flag.Func("color", "force or disable coloring of errors (`true`, `false`)", func(s string) error {
		ForceColorSetting = true

		switch s {
		case "", "true":
			Color = true
		case "false":
			Color = false
		default:
			return errors.New("invalid color setting, expected `true` or `false`")
		}

		return nil
	})
	flag.Func("colour", "force or disable colouring of errors, even if you're British (`true`, `false`)", func(s string) error {
		ForceColorSetting = true
		switch s {
		case "", "true":
			Color = true
		case "false":
			Color = false
		default:
			return errors.New("invalid colour setting, expected `true` or `false`")
		}

		return nil
	})

	var configDir string
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("AppData"); appData != "" {
			configDir = filepath.Join(appData, `Local\corgi`)
		}
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if home := os.Getenv("HOME"); home != "" {
			configDir = filepath.Join(home, ".config/corgi")
		}
	}

	var exePreferencesText string
	if configDir != "" {
		exePreferencesText = "\nthis does not affect preferences stored in `" + filepath.Join(configDir, "trusted_filters")
	}

	flag.Func("trust-filter", "trust these comma-separated executables to be run as filters"+exePreferencesText,
		func(s string) error {
			TrustedExecutables = append(TrustedExecutables, strings.Split(s, ",")...)
			return nil
		})
	flag.Func("trust-all-filters",
		"set to `i know this is dangerous` to allow running all executables as filters\n"+
			"only set this if you trust the file you are compiling or are running corgi in a secure environment (i.e. a container)",
		func(s string) error {
			if s == "i know this is dangerous" {
				TrustAllExecutables = true
				return nil
			}

			return fmt.Errorf("invalid value for `-trust-all-filters` flag, consult help (`-h`): %s", s)
		})

	flag.Parse()

	if showHelp {
		flag.Usage()
		return
	} else if showVersion {
		fmt.Println("corgi", version())
		return
	}

	if !PrecompileLibrary && Package == "" {
		Package = os.Getenv("GOPACKAGE")
		if Package == "" {
			wd, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "failed to get workdir\n"+
					"please set package manually using the `-package` flag")
				os.Exit(2)
			}

			Package = filepath.Base(wd)
		}
	}

	if OutFile != "" && UseStdout {
		fmt.Fprintln(os.Stderr, "conflicting flags -o and -stdout")
		os.Exit(2)
	}

	InFile = flag.Arg(0)
	if InFile == "" {
		if PrecompileLibrary {
			fmt.Fprintln(os.Stderr, "need directory to precompile")
			os.Exit(2)
		} else {
			var err error
			InData, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, "could not read stdin", err.Error())
				os.Exit(2)
			}

			if len(InData) == 0 {
				fmt.Fprintln(os.Stderr, "expected either input via stdin or a filepath as arg")
				os.Exit(2)
			}
		}
	}

	if OutFile != "" && PrecompileLibrary && InFile == "./..." {
		fmt.Fprintln(os.Stderr, "cannot use `-o` with `./...`")
		os.Exit(2)
	}

	if OutFile == "" && !UseStdout {
		if PrecompileLibrary {
			OutFile = filepath.Join(InFile, corgi.PrecompFileName)
		} else {
			if InFile == "" {
				fmt.Fprintln(os.Stderr, "need to specify output file, -o, or not use stdin")
				os.Exit(2)
			}

			OutFile = filepath.Base(flag.Arg(0)) + ".go"
		}
	}

	if GoExecPath == "" {
		if goroot := os.Getenv("GOROOT"); goroot != "" {
			GoExecPath = filepath.Join(goroot, "bin", "go")
		}
	}
}

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "This is the compiler for the corgi template language.")
	fmt.Fprintln(flag.CommandLine.Output(), "https://github.com/mavolin/corgi")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Usage: corgi [options] [FILE]")
	fmt.Fprintln(flag.CommandLine.Output(), "Usage: corgi [options] -lib DIR")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Input may be passed through stdin, however, this will disable loading of the file's dir library.")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "If -lib is specified, the FILE/DIR argument is mandatory.")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Additionally, the special ./.. argument is allowed, which recursively iterates")
	fmt.Fprintln(flag.CommandLine.Output(), "through pwd and its subdirectories and pre-compiles all of those containing corgi")
	fmt.Fprintln(flag.CommandLine.Output(), "library files.")
	fmt.Fprintln(flag.CommandLine.Output(), "If ./... is used, the -o flag has no effect and the precompiled files will be")
	fmt.Fprintln(flag.CommandLine.Output(), "placed directly into the respective directories.")
	fmt.Fprintln(flag.CommandLine.Output())
	flag.PrintDefaults()
}

func version() string {
	ver := meta.Version
	if meta.Commit != meta.UnknownCommit {
		ver += " (" + meta.Commit + ")"
	}

	return ver
}
