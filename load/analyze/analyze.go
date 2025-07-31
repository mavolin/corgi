// Package analyze aggregates some contextual analysis for easy access later on.
//
// This is necessary, since corgi is very context-sensitive (mostly because of
// &-attributes).
// The analyzer analyzes components and component calls to determine, amongst
// other things, if they write (top-level) attributes and if &-attributes can
// be placed after a component call.
// You can have a look at [file.Component] and [file.ComponentCall] to see all
// the information that is gathered.
//
// Package analyze also collects the state variable placed throughout the
// package for easier access.
//
// Since the results of these analyses are needed quite a bit, especially
// during validation, they are centralized here, reducing duplication and
// potential errors.
package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/load/analyze/internal/analyze"
	"github.com/mavolin/corgi/v2/load/analyze/internal/check"
	"github.com/mavolin/corgi/v2/load/analyze/internal/context"
)

// todo: contextual escapes (e.g. url): don't allow strings with interpolation as part of a larger expression, to avoid confusion (?) (maybe only if not in parentheses?)
// todo: set ElementSpec.Analyzed
// todo: set ElementSpec.Type
// todo: set AttributeReference.Analyzed
// todo: no elements in script/style https://html.spec.whatwg.org/multipage/syntax.html#elements-2:raw-text-elements-3
// todo: no elements in textarea/title https://html.spec.whatwg.org/multipage/syntax.html#elements-2:escapable-raw-text-elements-3
// todo: no component call interpolations that can't write top-level attrs, but do
// todo: no attr, except class, is written to twice
// todo: html.Once is never called with primitives
// todo: interpolation inside script/style using zero coalescing always have a default
// todo: make sure element definition aliases are not circular
// todo: either all blocks have a default or none
// todo: ccs in cc bodies only attach attributes, no content
// todo: no top-level attributes
// todo: top-level and-placeholders not filled, if not in element
// todo: only attr in args, if component has and placeholder
// todo: attributes inside withs
// todo: top-level attributes through block defaults
// todo: use of & after writing to body
// todo: use of & in loop that writes to body
// todo: interpolated comp must not write attrs
// todo: comp required args set
// todo: type infer
// todo: comp-call attributes only write text
// todo: analyze attr type
// todo: typed attribute values for attribute literals that are unsafe, must not carry the attribute name in brackets
// todo: ast.Types that are parsed as AttributeType must carry attribute name in brackets
// todo: analyze: attr type: a boolean attr without explicit value (just name) that has no definition -> not allowed
// todo: block nested inside itself
// todo: component block required
// todo: component first permanent and placeholder
// todo: component: block: first and placeholder
// todo: attr ref: element
// todo: attr ref: rule
// todo: attr ref: type
// todo: element spec: type
// todo: cc interpolation arg must only write text
// todo: comp interpolation must only write text
// todo: block is top-level

type Options struct {
	// Logger is the logger used by the analyzer.
	//
	// Default: no logging
	Logger *slog.Logger
}

func (o *Options) applyDefaults() {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
}

// Analyze fills the fields in the package's and the package's file's symbols
// that are marked as analyzer fields.
//
// The package must be linked.
//
// If it returns an error, it is always of type [fileerr.List].
func Analyze(p *file.Package, o Options) diagnostic.List {
	if !p.Linked {
		panic("Analyze called with unlinked package")
	}

	o.applyDefaults()

	logger := o.Logger.With(
		slog.String("module", p.Module),
		slog.String("path_in_module", p.PathInModule),
	)

	p.Analyzed = true
	ctx := context.New(p, logger)
	analyze.Analyze(ctx)
	check.Check(ctx)
	return ctx.Diagnostics()
}
