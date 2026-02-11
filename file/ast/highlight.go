package ast

// Highlight returns the canonical start and end positions for highlighting
// the passed node.
// Highlight checks if the given node implements the [Highlighter] interface,
// and if so, uses those positions.
// Otherwise, it defaults to using the node's Start and End positions.
//
// Refer to [Highlighter] for the benefits of using the returned range instead
// of the default node.Start(), node.End() interval.
func Highlight(n Node) (start, end Position) {
	if h, ok := n.(Highlighter); ok {
		return h.Highlight()
	}
	return n.Start(), n.End()
}

// Highlighter is an interface that can be implemented by nodes to supply a
// custom highlighting range, instead of using the default
// [node.Start(), node.End()) interval.
//
// Nodes with a body, or other nodes that can get excessively long spans,
// should implement this interface to return a more concise highlighting range
// that focuses on the most relevant part of the node.
// This is done to prevent diagnostics from growing unnecessarily large.
type Highlighter interface {
	// Highlight returns the start and end positions for highlighting.
	//
	// Start is inclusive, end is exclusive.
	Highlight() (start, end Position)
}
