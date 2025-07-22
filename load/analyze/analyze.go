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
// todo: aggregate component blocks
// todo: required block overwritten by alias
// todo: aliases must be in same package
// todo: component block required
// todo: component first permanent and placeholder
// todo: component: block: first and placeholder
// todo: component call: first and
// todo: attr ref: element
// todo: attr ref: rule
// todo: attr ref: type
// todo: element spec: type
// todo: cc interpolation arg must only write text
// todo: comp interpolation must only write text
// todo: block is top-level
// todo: nested typed attribute values
// todo: and with and placeholder in alias component call (and placeholder gets inherited anyway, or should it?)
// todo: ast.AttributeType as ParsedType only allowed on component call parameter
// todo: scope type assertions don't work on underscore blocks
// todo: block annotations need to be corrected
// todo: file.(Block|With).Name correctly used (i.e. empty for default block)
// todo: looped with considers only direct children
// todo: extend as part of body
// todo: store param field on file.ComponentArgument
// todo: link/linkattribute_reference.go:122
// todo: block function needs to also accept no ident
// todo: check/misc_scope: todos

// todo: should withs include underscore block shorthand?? prob yes

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

// Analyze fills the package's State, and the remaining fields not set by
// package link in the package's Components and ComponentCalls.
//
// If it returns an error, it is always of type [fileerr.List].
func Analyze(p *file.Package, o Options) diagnostic.List {
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
