package gocmd

import (
	"encoding/json"
	"errors"
)

type DownloadModule struct {
	Path     string // module path
	Query    string // version query corresponding to this version
	Version  string // module version
	Error    string // error loading module
	Info     string // absolute path to cached .info file
	GoMod    string // absolute path to cached .mod file
	Zip      string // absolute path to cached .zip file
	Dir      string // absolute path to cached source root directory
	Sum      string // checksum for path, version (as in go.sum)
	GoModSum string // checksum for go.mod (as in go.sum)
	Origin   any    // provenance of module
	Reuse    bool   // reuse of old module info is safe
}

func (c *Cmd) DownloadMod(path string) (*DownloadModule, error) {
	data, cmdErr := c.command("mod", "download", "-json", path)
	// depending on the err, ^ will return a json error message

	var mod DownloadModule
	if err := json.Unmarshal(data, &mod); err != nil {
		if cmdErr != nil {
			return nil, cmdErr
		}

		return nil, err
	}
	if mod.Error != "" {
		return &mod, errors.New(mod.Error)
	}

	return &mod, nil
}
