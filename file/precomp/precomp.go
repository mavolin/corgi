// Package precomp allows encoding and decoding of precompiled libraries.
package precomp

import (
	"io"

	"github.com/tinylib/msgp/msgp"

	cfile "github.com/mavolin/corgi/file"
)

func Encode(w io.Writer, l *cfile.Library) error {
	if !l.Precompiled {
		panic("precomp.Encode: trying to encode non-precompiled library: " + l.Module + "/" + l.PathInModule)
	}

	lw, err := newLibrary(l)
	if err != nil {
		return err
	}

	return msgp.Encode(w, lw)
}

func Marshal(l *cfile.Library) ([]byte, error) {
	lw, err := newLibrary(l)
	if err != nil {
		return nil, err
	}

	return lw.MarshalMsg(nil)
}

func Decode(r io.Reader) (*cfile.Library, error) {
	var l library
	if err := msgp.Decode(r, &l); err != nil {
		return nil, err
	}

	return l.toFile(), nil
}

func Unmarshal(in []byte) (*cfile.Library, error) {
	var l library
	if _, err := l.UnmarshalMsg(in); err != nil {
		return nil, err
	}

	return l.toFile(), nil
}
