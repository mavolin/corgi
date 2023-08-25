package std

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/precomp"
)

//go:generate corgi -lib ./...

// html/lib.precorgi strings/lib.precorgi
//
//go:embed fmt/lib.precorgi html/lib.precorgi strings/lib.precorgi
var libFiles embed.FS

var Lib map[string]*file.Library

func init() {
	Lib = make(map[string]*file.Library, 1)

	_ = fs.WalkDir(libFiles, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "std: %s: failed to walk embedded stdlib: %s", p, err.Error())
			return nil
		}

		if d.Name() != "lib.precorgi" {
			return nil
		}

		f, err := libFiles.Open(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "std: %s: failed to open embedded precompiled stdlib: %s", p, err.Error())
			return nil
		}

		defer func() {
			err := f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "std: %s: failed to close embedded precompiled stdlib: %s", p, err.Error())
			}
		}()

		lib, err := precomp.Decode(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "std: %s: failed to decode embedded precompiled stdlib: %s", p, err.Error())
			return nil
		}

		p = path.Dir(strings.TrimPrefix(p, "./"))
		lib.Module = ""
		lib.PathInModule = p
		Lib[p] = lib
		return nil
	})
}
