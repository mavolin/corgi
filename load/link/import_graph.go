package link

import (
	"context"
	"sync"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

// The importGraph is a directed graph representing the import relationships
// between packages.
// It is used both for caching and for cycle detection.
//
// Each time a package A wants to load another package B, it needs to add an
// edge A -> B to the graph.
//
// If no vertex for package B exist yet, A now bears the responsibility to load
// B and store its result in the graph.
//
// If a vertex for B already exists, A needs to verify that there is no path
// from B back to A, i.e. no import cycle exists.
// Only then can it wait for B's "owner" to finish loading B and use the result.
// If it didn't check for cycles, A would deadlock waiting for B, as B is
// already waiting for A to finish loading.
//
// The graph is concurrently safe and must be shared from the root down to all
// imports.
type (
	importGraph struct {
		mu       sync.Mutex
		packages map[file.CorgiImportPath]*packageNode // all vertices in the graph
	}

	packageNode struct {
		loaded      chan struct{} // closed when the package has been loaded
		pkg         *file.Package
		diagnostics diagnostic.List
		err         error

		imports map[file.CorgiImportPath]*packageNode // edges to other vertices
	}
)

type importGraphKey struct{}

func importGraphFromContext(ctx context.Context) (context.Context, *importGraph) {
	g, _ := ctx.Value(importGraphKey{}).(*importGraph)
	if g == nil {
		g = newImportGraph()
		ctx = context.WithValue(ctx, importGraphKey{}, g)
	}
	return ctx, g
}

func newImportGraph() *importGraph {
	return &importGraph{
		packages: make(map[file.CorgiImportPath]*packageNode),
	}
}

func (g *importGraph) AddImport(
	parent *file.Package, impPath file.CorgiImportPath, compute func() (*file.Package, diagnostic.List, error),
) (cycle []file.CorgiImportPath) {
	g.mu.Lock()
	defer g.mu.Unlock()

	parentNode := g.packages[parent.CorgiImportPath]
	if parentNode == nil {
		parentNode = &packageNode{
			loaded:  make(chan struct{}),
			imports: make(map[file.CorgiImportPath]*packageNode),
		}
		close(parentNode.loaded) // parent is already loaded
		g.packages[parent.CorgiImportPath] = parentNode
	}

	childNode := g.packages[impPath]
	if childNode == nil {
		childNode = &packageNode{
			loaded:  make(chan struct{}),
			imports: make(map[file.CorgiImportPath]*packageNode),
		}
		g.packages[impPath] = childNode
		parentNode.imports[impPath] = childNode

		go func() {
			childNode.pkg, childNode.diagnostics, childNode.err = compute()
			close(childNode.loaded)
		}()

		return nil
	}

	if parentNode == childNode {
		// This should've been caught by the self-import check.
		panic("importGraph: tried to add self-loop")
	}

	parentNode.imports[impPath] = childNode

	if path := g.findPath(childNode, parentNode); path != nil {
		// cycle detected
		return path
	}

	return nil
}

func (g *importGraph) AwaitImport(impPath file.CorgiImportPath) (*file.Package, diagnostic.List, error) {
	g.mu.Lock()
	node := g.packages[impPath]
	g.mu.Unlock()

	if node == nil {
		// This should never happen.
		panic("importGraph: tried to await unknown package")
	}

	<-node.loaded
	return node.pkg, node.diagnostics, node.err
}

// findPath returns the shortest path from 'from' to 'to' if one exists.
//
// from and to must not be equal.
func (g *importGraph) findPath(from, to *packageNode) []file.CorgiImportPath {
	type queueItem struct {
		node  *packageNode
		cycle []file.CorgiImportPath
	}

	// Use BFS to find the shortest path.
	queue := []queueItem{{from, make([]file.CorgiImportPath, 0, 8)}}
	seen := make(map[*packageNode]bool)
	seen[from] = true
	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]

		for childImportPath, child := range parent.node.imports {
			if seen[child] {
				continue
			}
			seen[child] = true

			path := make([]file.CorgiImportPath, len(parent.cycle)+1)
			copy(path, parent.cycle)
			path[len(parent.cycle)] = childImportPath

			if child == to {
				return path
			}

			queue = append(queue, queueItem{child, path})
		}
	}
	return nil
}
