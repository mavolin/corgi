package attrtype

import (
	"regexp"
	"strings"
)

// https://htmx.org/extensions/response-targets/
var htmxResponseTargetsRegexp = regexp.MustCompile(`^hx-target-(?:error|\*|[1-5](?:\d\d|\d?\*))$`)

// HTMX is a [Func] that returns the [Type] for HTMX attributes as defined in
// the [HTMX Attribute Reference].
//
// [HTMX Attribute Reference]: https://htmx.org/docs/#attributes
func HTMX(_, attr string) Type {
	switch attr {
	//
	// https://htmx.org/reference/#attributes

	case "hx-boost":
		return Text
	case "hx-get":
		return URL
	case "hx-post":
		return URL
	case "hx-on":
		return JS
	case "hx-push-url":
		return URL
	case "hx-select":
		return Text
	case "hx-select-oob":
		return Text
	case "hx-swap":
		return Text
	case "hx-swap-oob":
		return Text
	case "hx-target":
		return Text
	case "hx-trigger":
		return JS // not perfect, but when else would you interpolate
	case "hx-vals":
		return JS

	//
	// https://htmx.org/reference/#attributes-additional

	case "hx-confirm":
		return Text
	case "hx-delete":
		return URL
	case "hx-disable":
		return Text
	case "hx-disabled-elt":
		return Text
	case "hx-disinherit":
		return Text
	case "hx-encoding":
		return Unsafe
	case "hx-ext":
		return Text
	case "hx-headers":
		return JS
	case "hx-history":
		return Text
	case "hx-history-elt":
		return Text
	case "hx-include":
		return Text
	case "hx-indicator":
		return Text
	case "hx-params":
		return Unsafe
	case "hx-patch":
		return URL
	case "hx-presence":
		return Text
	case "hx-prompt":
		return Text
	case "hx-put":
		return URL
	case "hx-replace-url":
		return URL
	case "hx-request":
		return Unsafe
	case "hx-sync":
		return Text
	case "hx-validate":
		return Unsafe
	case "hx-vars":
		return Unsafe // hx-vars is deprecated, use hx-vals instead

	// https://htmx.org/extensions/class-tools
	case "classes":
		return Text

	// https://htmx.org/extensions/client-side-templates
	case "mustache-template":
		return Text
	case "handlebars-template":
		return Text
	case "nunjucks-template":
		return Text
	case "xslt-template":
		return Text

	// https://htmx.org/extensions/include-vals
	case "include-vals":
		return JS

	// https://htmx.org/extensions/loading-states
	case "data-loading":
		return Text
	case "data-loading-class":
		return Text
	case "data-loading-class-remove":
		return Text
	case "data-loading-disable":
		return Text
	case "data-loading-aria-busy":
		return Text
	case "data-loading-delay":
		return Text
	case "data-loading-target":
		return Text
	case "data-loading-path":
		return URL
	case "data-loading-states":
		return Text

	// https://htmx.org/extensions/path-deps
	case "path-deps":
		return URL

	// https://htmx.org/extensions/preload
	case "preload":
		return Text

	// https://htmx.org/extensions/remove-me
	case "remove-me":
		return Text

	// https://htmx.org/extensions/server-sent-events
	case "sse-connect":
		return ResourceURL
	case "sse-swap":
		return Text

	// https://htmx.org/extensions/web-sockets/
	case "ws-connect":
		return ResourceURL
	case "ws-send":
		return Text
	}

	if strings.HasPrefix(attr, "hx-on:") {
		return JS
	}

	if htmxResponseTargetsRegexp.MatchString(attr) {
		return Text
	}

	return Unknown
}

var _ Func = HTMX
