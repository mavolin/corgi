package file

import (
	"strings"

	"github.com/mavolin/corgi/v2/internal/assert"
	"github.com/mavolin/corgi/v2/internal/meta"
	"golang.org/x/mod/module"
)

const (
	EscapeImport  GoImportPath = meta.Module + "/escape"
	SafeImport    GoImportPath = EscapeImport + "/safe"
	RuntimeImport GoImportPath = meta.Module + "/runtime"
)

func init() { //nolint:gochecknoinits
	assert.NoError(SafeImport.CheckValid(), "SafeImport invalid")
	assert.NoError(RuntimeImport.CheckValid(), "RuntimeImport invalid")
	assert.NoError(EscapeImport.CheckValid(), "EscapeImport invalid")
}

type (
	// Qualifier is the qualifier for an import.
	Qualifier string
	// ModulePath is the path/name of a Go module.
	ModulePath string
	// PackagePath is a path to a package relative to the root of a Go module, with no
	// leading './' or '/'.
	PackagePath string
	// Path is a path relative to the root of a Go module, with no
	// leading './' or '/'.
	Path string
	// Name is the name of a file.
	Name            string
	CorgiImportPath string
	GoImportPath    string

	// Type is a Go type, named or unnamed, meaningful only in the package it
	// is used in.
	Type string
	// Identifier refers to a Go symbol defined in a package.
	Identifier string

	// CanonicalAttributeName is the canonical name of an HTML attribute, i.e.
	// the ascii-lowercase name.
	//
	// Unlike elements, attributes may collide with keywords, so the canonical
	// form is always the same as the idiomatic form.
	CanonicalAttributeName string
	// CanonicalElementName is the canonical name of an HTML element, i.e. the
	// ascii-lowercase name.
	CanonicalElementName string

	// CanonicalQualifiableElementName is the canonical partial name of an HTML element, used
	// in conjunction with a qualifier to form a valid qualified element name.
	CanonicalQualifiableElementName string
	// CanonicalQualifiableAttributeName is the canonical partial name of an HTML
	// attribute, used in conjunction with a qualifier to form a valid
	// qualified attribute name.
	CanonicalQualifiableAttributeName string
)

type canonicalName interface {
	CanonicalAttributeName | CanonicalElementName |
		CanonicalQualifiableElementName | CanonicalQualifiableAttributeName
}

// Canonicalize returns the canonical form of the given name, i.e. the
// ascii-lowercase form, as the given type N.
//
// See: https://html.spec.whatwg.org/multipage/parsing.html#tag-name-state and
// https://html.spec.whatwg.org/multipage/parsing.html#attribute-name-state
func Canonicalize[N canonicalName](name string) N {
	var b strings.Builder
	for i, r := range name {
		if r < 'A' || r > 'Z' {
			continue
		}

		b.Grow(len(name))
		b.WriteString(name[:i])
		b.WriteRune((r - 'A') + 'a')
		for _, r := range name[i+1:] {
			if r >= 'A' && r <= 'Z' {
				b.WriteRune((r - 'A') + 'a')
			} else {
				b.WriteRune(r)
			}
		}
	}

	if b.Len() == 0 {
		return N(name)
	}
	return N(b.String())
}

func (q Qualifier) QualifiedType(t Type) Type {
	return Type(q) + "." + t
}

func (id Identifier) Exported() bool {
	return len(id) > 0 && 'A' <= id[0] && id[0] <= 'Z'
}

func (m ModulePath) ImportPathFor(p PackagePath) GoImportPath {
	return GoImportPath(string(m) + "/" + string(p))
}

func (pp PackagePath) FilePath(n Name) Path {
	return Path(string(pp) + "/" + string(n))
}

// CheckValid reports whether the import path is syntactically valid.
func (p CorgiImportPath) CheckValid() error {
	return module.CheckImportPath(string(p))
}

// CheckValid reports whether the import path is syntactically valid.
func (p GoImportPath) CheckValid() error {
	return module.CheckImportPath(string(p))
}
