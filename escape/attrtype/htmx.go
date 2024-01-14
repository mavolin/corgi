package attrtype

import (
	"regexp"
	"strings"
)

// HTMX is a [Func] that returns the [Type] for HTMX attributes as defined in
// the [HTMX Attribute Reference].
//
// [HTMX Attribute Reference]: https://htmx.org/docs/#attributes
func HTMX(_, attr string) Type {
	if t := htmx[attr]; t.IsValid() {
		return t
	}

	if strings.HasPrefix(attr, "hx-on:") {
		return JS
	}

	if htmxResponseTargetsRegexp.MatchString(attr) {
		return Plain
	}

	return Unknown
}

var _ Func = HTMX

var (
	htmx = map[string]Type{
		// https://htmx.org/reference/#attributes
		"hx-boost":      Plain,
		"hx-get":        URL,
		"hx-post":       URL,
		"hx-on":         JS,
		"hx-push-url":   URL,
		"hx-select":     Plain,
		"hx-select-oob": Plain,
		"hx-swap":       Plain,
		"hx-swap-oob":   Plain,
		"hx-target":     Plain,
		"hx-trigger":    JS, // not perfect, but when else would you interpolate
		"hx-vals":       JS,

		// https://htmx.org/reference/#attributes-additional
		"hx-confirm":      Plain,
		"hx-delete":       URL,
		"hx-disable":      Plain,
		"hx-disabled-elt": Plain,
		"hx-disinherit":   Plain,
		"hx-encoding":     Unsafe,
		"hx-ext":          Plain,
		"hx-headers":      JS,
		"hx-history":      Plain,
		"hx-history-elt":  Plain,
		"hx-include":      Plain,
		"hx-indicator":    Plain,
		"hx-params":       Unsafe,
		"hx-patch":        URL,
		"hx-presence":     Plain,
		"hx-prompt":       Plain,
		"hx-put":          URL,
		"hx-replace-url":  URL,
		"hx-request":      Unsafe,
		"hx-sync":         Plain,
		"hx-validate":     Unsafe,
		"hx-vars":         Unsafe, // hx-vars is deprecated, use hx-vals instead

		// https://htmx.org/extensions/class-tools
		"classes": Plain,

		// https://htmx.org/extensions/client-side-templates
		"mustache-template":   Plain,
		"handlebars-template": Plain,
		"nunjucks-template":   Plain,
		"xslt-template":       Plain,

		// https://htmx.org/extensions/include-vals
		"include-vals": JS,

		// https://htmx.org/extensions/loading-states
		"data-loading":              Plain,
		"data-loading-class":        Plain,
		"data-loading-class-remove": Plain,
		"data-loading-disable":      Plain,
		"data-loading-aria-busy":    Plain,
		"data-loading-delay":        Plain,
		"data-loading-target":       Plain,
		"data-loading-path":         URL,
		"data-loading-states":       Plain,

		// https://htmx.org/extensions/path-deps
		"path-deps": URL,

		// https://htmx.org/extensions/preload
		"preload": Plain,

		// https://htmx.org/extensions/remove-me
		"remove-me": Plain,

		// https://htmx.org/extensions/server-sent-events
		"sse-connect": ResourceURL,
		"sse-swap":    Plain,

		// https://htmx.org/extensions/web-sockets/
		"ws-connect": ResourceURL,
		"ws-send":    Plain,
	}

	// https://htmx.org/extensions/response-targets/
	htmxResponseTargetsRegexp = regexp.MustCompile(`^hx-target-(?:error|\*|[1-5](?:\d\d|\d?\*))$`)
)
