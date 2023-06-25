// Command corgidebug is a utility for debugging problems, when the compiler
// doesn't do what it should.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/k0kubun/pp"
	"github.com/mattn/go-isatty"

	"github.com/mavolin/corgi/corgierr"
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

	pp.Println(f)

	fmt.Printf("\nparse took %s\n\n", time.Since(now))

	if err != nil {
		errs := err.(corgierr.List)
		for _, err := range errs {
			fmt.Println("")
			fmt.Println(err.Pretty(corgierr.PrettyOptions{
				Colored: isatty.IsTerminal(os.Stdout.Fd()),
			}))
		}

		return
	}

	now = time.Now()
	err = validate.UseNamespaces(f)
	fmt.Printf("use namespace validation took %s\n\n", time.Since(now))
	if err != nil {
		errs := err.(corgierr.List)
		for _, err := range errs {
			fmt.Println("")
			fmt.Println(err.Pretty(corgierr.PrettyOptions{
				Colored: isatty.IsTerminal(os.Stdout.Fd()),
			}))
		}

		return
	}

	now = time.Now()
	err = link.New(nil).Link(f)
	fmt.Printf("linking took %s\n\n", time.Since(now))
	if err != nil {
		errs := err.(corgierr.List)
		for _, err := range errs {
			fmt.Println("")
			fmt.Println(err.Pretty(corgierr.PrettyOptions{
				Colored: isatty.IsTerminal(os.Stdout.Fd()),
			}))
		}

		return
	}

	now = time.Now()
	err = validate.File(f)
	fmt.Printf("validation took %s\n\n", time.Since(now))
	if err != nil {
		errs := err.(corgierr.List)
		for _, err := range errs {
			fmt.Println("")
			fmt.Println(err.Pretty(corgierr.PrettyOptions{
				Colored: isatty.IsTerminal(os.Stdout.Fd()),
			}))
		}

		return
	}
}
