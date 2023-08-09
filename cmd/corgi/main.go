package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"golang.org/x/exp/slog"

	"github.com/mavolin/corgi"
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/write"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run() error {
	loadOpts := corgi.LoadOptions{GoExecPath: GoExecPath}
	if Verbose {
		loadOpts.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	if PrecompileLibrary {
		return writeLibraries(loadOpts)
	}

	return writeFile(loadOpts)
}

func writeFile(loadOpts corgi.LoadOptions) error {
	var f *file.File
	var err error
	if InFile != "" {
		f, err = corgi.LoadMain(InFile, loadOpts)
	} else {
		f, err = corgi.LoadMainData(InData, loadOpts)
	}

	if err != nil {
		var mainMod string
		if f != nil {
			mainMod = f.Module
		}
		writeErrs(err, mainMod)
		return nil
	}

	var out io.Writer
	closeOut := func() error { return nil }
	if UseStdout {
		out = os.Stdout
	} else {
		fout, err := os.Create(OutFile)
		if err != nil {
			return fmt.Errorf("could not create output file: %w", err)
		}
		closeOut = fout.Close
		out = fout
	}

	prettyOut := out
	prettyClose := func() error { return nil }

	var goImportsErr error
	var goImportsDone <-chan struct{}

	if !NoGoImports {
		done := make(chan struct{})
		goImportsDone = done

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		goimports := exec.CommandContext(ctx, "goimports")
		goimports.Stdout = out
		stderr := bytes.NewBuffer(make([]byte, 0, 256))
		goimports.Stderr = stderr
		pipe, err := goimports.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to pipe generated file into goimports: %w", err)
		}

		prettyOut = pipe
		prettyClose = pipe.Close

		go func() {
			goImportsErr = goimports.Run()
			if stderr.Len() > 0 {
				goImportsErr = errors.New(stderr.String())
			}
			close(done)
		}()
	}

	w := write.New(write.Options{
		AllowedFilters:  TrustedFilters,
		AllowAllFilters: TrustAllFilters,
		CLI:             true,
		CorgierrPretty:  prettyOptions(f.Module),
		Debug:           Debug,
	})

	if err := w.GenerateFile(prettyOut, Package, f); err != nil {
		return err
	}

	if err := prettyClose(); err != nil {
		return fmt.Errorf("close goimports pipe: %w", err)
	}

	if !NoGoImports {
		<-goImportsDone
		if goImportsErr != nil {
			// goimport's error probably contains line/col info, so generate the
			// file again, but this time directly
			w := write.New(write.Options{Debug: Debug})
			_ = w.GenerateFile(out, Package, f)

			return fmt.Errorf("failed to run goimports:\n"+
				"\tyou probably have an erroneous expression in your corgi file,\n"+
				"\tor need to re-precompile a library: "+
				"\t\t%w", goImportsErr)
		}
	}

	if err := closeOut(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}

	return nil
}

func writeLibraries(loadOpts corgi.LoadOptions) error {
	loadOpts.NoPrecompile = true

	if InFile != "./..." {
		return writeLibrary(InFile, OutFile, loadOpts)
	}

	return fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return writeLibrary(path, filepath.Join(path, corgi.PrecompFileName), loadOpts)
		}

		return nil
	})
}

func writeLibrary(path, outPath string, loadOpts corgi.LoadOptions) error {
	lib, err := corgi.LoadLibrary(path, loadOpts)
	if errors.Is(err, corgi.ErrExists) {
		return nil
	} else if err != nil {
		var mainMod string
		if lib != nil {
			mainMod = lib.Module
		}
		writeErrs(err, mainMod)
		return nil
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("%s: failed to open output: %w", path, outPath)
	}

	w := write.New(write.Options{
		AllowedFilters:  TrustedFilters,
		AllowAllFilters: TrustAllFilters,
		CLI:             true,
		CorgierrPretty:  prettyOptions(lib.Module),
		Debug:           Debug,
	})

	if err := w.PrecompileLibrary(out, lib); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}

	return nil
}

func writeErrs(err error, mainMod string) {
	prettyOpts := prettyOptions(mainMod)

	var clerr corgierr.List
	if errors.As(err, &clerr) {
		fmt.Println(clerr.Pretty(prettyOpts))
		os.Exit(1)
	}

	var cerr *corgierr.Error
	if errors.As(err, &cerr) {
		fmt.Println(clerr.Pretty(prettyOpts))
		os.Exit(1)
	}

	fmt.Println(err)
	os.Exit(1)
}

func prettyOptions(mainMod string) corgierr.PrettyOptions {
	var prettyOpts corgierr.PrettyOptions

	color.Set()

	prettyOpts.Colored = Color
	if !ForceColorSetting {
		prettyOpts.Colored = os.Getenv("TERM") != "dumb" &&
			(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))
	}

	if IsGoGenerate {
		prettyOpts.FileNamePrinter = func(f *file.File) string {
			if f.Module == mainMod {
				return filepath.FromSlash(f.PathInModule)
			}

			return path.Join(f.Module, f.PathInModule)
		}
	} else {
		wd, err := os.Getwd()
		if err == nil { // IS nil
			// print the name of the file relative to the current wd
			prettyOpts.FileNamePrinter = func(f *file.File) string {
				rel, relErr := filepath.Rel(wd, f.AbsolutePath)
				if relErr != nil {
					return rel
				}

				return f.Name
			}
		}
	}

	return prettyOpts
}
