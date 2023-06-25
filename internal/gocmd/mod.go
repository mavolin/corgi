package gocmd

import (
	"encoding/json"
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
	data, err := c.command("mod", "download", "-json", path)
	if err != nil {
		return nil, err
	}

	var mod DownloadModule
	return &mod, json.Unmarshal(data, &mod)
}
