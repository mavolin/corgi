// Command corgidebug is a utility for debugging problems when the compiler
// doesn't do what it should.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/k0kubun/pp"
	"github.com/mattn/go-isatty"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file/typeinfer"
	"github.com/mavolin/corgi/link"
	"github.com/mavolin/corgi/parse"
	"github.com/mavolin/corgi/validate"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	now := time.Now()
	f, err := parse.Parse(data)
	//nolint:govet // we actually don't want time.Since to be deferred
	defer fmt.Printf("\nparse took %s\n", time.Since(now))

	if f != nil {
		f.Name = os.Args[1]
	}
	if err != nil {
		pp.Println(f)

		errs := err.(corgierr.List) //nolint:errorlint
		for _, err := range errs {
			fmt.Println("")
			fmt.Println(err.Pretty(corgierr.PrettyOptions{
				Colored: isatty.IsTerminal(os.Stdout.Fd()),
			}))
		}

		return
	}

	now = time.Now()
	err = validate.PreLink(f)
	defer fmt.Printf("\nuse namespace validation took %s", time.Since(now)) //nolint:govet // we actually don't want time.Since to be deferred
	if err != nil {
		errs := corgierr.As(err)
		fmt.Println("")
		fmt.Println(errs.Pretty(corgierr.PrettyOptions{
			Colored: isatty.IsTerminal(os.Stdout.Fd()),
		}))

		return
	}

	now = time.Now()
	typeinfer.Scope(f.Scope)
	defer fmt.Printf("\nmixin param type infer took %s", time.Since(now)) //nolint:govet // we actually don't want time.Since to be deferred

	now = time.Now()
	err = link.New(nil).LinkFile(f)
	linkDura := time.Since(now)

	pp.Println(f)
	defer fmt.Printf("\nlinking took %s", linkDura)
	if err != nil {
		errs := err.(corgierr.List) //nolint:errorlint
		fmt.Println("")
		fmt.Println(errs.Pretty(corgierr.PrettyOptions{
			Colored: isatty.IsTerminal(os.Stdout.Fd()),
		}))

		return
	}

	now = time.Now()
	err = validate.File(f)
	//nolint:govet // we actually don't want time.Since to be deferred
	defer fmt.Printf("\nvalidation took %s", time.Since(now))
	if err != nil {
		errs := err.(corgierr.List) //nolint:errorlint
		fmt.Println("")
		fmt.Println(errs.Pretty(corgierr.PrettyOptions{
			Colored: isatty.IsTerminal(os.Stdout.Fd()),
		}))

		return
	}
}
