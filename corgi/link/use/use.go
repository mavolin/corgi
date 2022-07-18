// Package use provides a NamespaceChecker that checks for namespace collisions
// in use directives.
package use

import "github.com/mavolin/corgi/corgi/file"

// NamespaceChecker checks for namespace collisions in use directives.
type NamespaceChecker struct {
	f file.File
}

// NewNamespaceChecker creates a new NamespaceChecker that checks the given
// file.
func NewNamespaceChecker(f file.File) *NamespaceChecker {
	return &NamespaceChecker{f: f}
}

// Check checks that there are no two use directives that use the same
// namespace.
func (c *NamespaceChecker) Check() error {
	for i, use := range c.f.Uses {
		if use.Namespace == "." || use.Namespace == "_" {
			continue
		}

		for _, cmp := range c.f.Uses[i+1:] {
			if use.Namespace == cmp.Namespace {
				return &NamespaceError{
					Source:    c.f.Source,
					File:      c.f.Name,
					Line:      use.Line,
					OtherLine: cmp.Line,
					Namespace: string(use.Namespace),
				}
			}
		}
	}

	return nil
}
