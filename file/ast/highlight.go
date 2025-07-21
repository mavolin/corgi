package ast

// Highlight returns the canonical start and end positions for highlighting
// the passed node.
// It is a convenience function that can be used to highlight nodes, while not
// needing to worry about excessively long spanning highlights for typically
// bigger nodes like components or component calls.
//
// The start position is inclusive, the end position is exclusive.
func Highlight(n Node) (start, end Position) {
	if h, ok := n.(Highlighter); ok {
		return h.Highlight()
	}
	return n.Start(), n.End()
}

// Highlighter is an interface that can be implemented by nodes to supply a
// custom highlighting range, instead of using the default
// [node.Start(), node.End()) interval.
type Highlighter interface {
	// Highlight returns the start and end positions for highlighting.
	//
	// Start is inclusive, end is exclusive.
	Highlight() (start, end Position)
}
