package app

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mavolin/corgi/corgi"
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/internal/meta"
	"github.com/mavolin/corgi/writer"
)

var log *zap.SugaredLogger

func init() {
	c := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     ".",
			LevelKey:       ".",
			TimeKey:        ".",
			NameKey:        ".",
			CallerKey:      zapcore.OmitKey,
			FunctionKey:    zapcore.OmitKey,
			StacktraceKey:  zapcore.OmitKey,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := c.Build()
	if err != nil {
		panic(err)
	}

	log = logger.Sugar()
}

func Run(args []string) error {
	ver := meta.Version
	if meta.Commit != meta.UnknownCommit {
		ver += " (" + meta.Commit + ")"
	}

	app := &cli.App{
		Name:  "corgi",
		Usage: "Generate Go functions from corgi files",
		Description: "This is the compiler for the corgi template language.\n\n" +
			"https://github.com/mavolin/corgi",
		Version:   ver,
		ArgsUsage: "<input file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "package",
				Aliases:  []string{"p"},
				Usage:    "set the name of the package to generate into; not required when using go generate",
				EnvVars:  []string{"GOPACKAGE"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:      "resource",
				Aliases:   []string{"r"},
				Usage:     "add `DIR` to the list of resource directories",
				TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "ignorecorgi",
				Usage: "don't use the $projectRoot/corgi resource directory",
			},
			&cli.StringFlag{
				Name:        "filetype",
				Aliases:     []string{"t"},
				Usage:       "overwrite the file type of the file (html, xhtml, xml)",
				DefaultText: "",
				Value:       "",
			},
			&cli.StringFlag{
				Name:        "filename",
				Aliases:     []string{"f"},
				Usage:       "overwrite the name of the generated file",
				DefaultText: "corgi_file.corgi.go",
			},
			&cli.BoolFlag{
				Name:  "nofmt",
				Usage: "don't format the output and remove unused imports",
			},
			&cli.BoolFlag{
				Name:  "get",
				Usage: "go get github.com/mavolin/corgi before generating the function",
			},
		},
		HideHelpCommand:      true,
		EnableBashCompletion: true,
		Action:               run,
	}

	return app.Run(args)
}

func run(ctx *cli.Context) error {
	//goland:noinspection GoBoolExpressions
	if meta.Version == meta.DevelopVersion {
		log.Warn("you're running a development version of corgi, to get the stable release, " +
			"run `go install github.com/mavolin/corgi/cmd/corgi@latest`")
	}

	args, err := parseArgs(ctx)
	if err != nil {
		return err
	}

	if args.Get {
		goGetCorgi()
	}

	ph := corgi.File(".", args.File, args.FileContents)

	if args.FileType != file.TypeUnknown {
		ph.WithFileType(args.FileType)
	}

	for _, rs := range args.ResourceSources {
		ph.WithResourceSource(rs)
	}

	f, err := ph.Parse()
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	w := writer.New(f, args.Package)

	out, err := os.Create(args.OutputFile)
	if err != nil {
		return errors.Wrap(err, "could not create output file")
	}

	if err := w.Write(out); err != nil {
		return err
	}

	log.Info("generated ", args.OutputFile)

	if !args.NoFmt {
		goImports(args)
	}

	return out.Close()
}

func goGetCorgi() {
	log.Debug("generated functions import github.com/mavolin/corgi, I'm go getting it for you")

	goget := exec.Command("go", "get", "github.com/mavolin/corgi")
	goget.Stderr = os.Stderr

	if err := goget.Run(); err != nil {
		log.Error("couldn't go get corgi: ", err.Error(),
			"; please do it yourself if you haven't already: go get github.com/mavolin/corgi")
	}
}

func goImports(args *args) {
	goimports := exec.Command("goimports", "-w", args.OutputFile) //nolint:gosec

	err := goimports.Run()
	if err == nil {
		log.Debug("formatted output and removed unused imports, if any")
		return
	}

	if errors.Is(err, exec.ErrNotFound) {
		log.Error("goimports could not be found, but is needed to remove unused imports; " +
			"install using `go get golang.org/x/tools/cmd/goimports@latest`")
		return
	}

	log.Error("could not format output "+
		"(this could mean that there is an erroneous Go expression in your template): ",
		err.Error())
}
