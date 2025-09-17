package file

import "github.com/mavolin/corgi/v2/file/ast"

type State struct {
	//
	// BUILD SYMBOLS

	AST  *ast.StateSpec
	File *File
	// Index of the variable in the Names and Values slices.
	//
	// There might be gaps in the indexes:
	// If the parser cannot parse the name of a variable, it will be excluded
	// from the symbols.
	//
	// There are only as many indexes as there are names, excess values are
	// ignored.
	Index int

	//
	// ANALYZER

	// Analyzed indicates whether the State has been analyzed,
	// albeit with errors.
	Analyzed bool

	// The InferredType of this value, if there is no explicit type.
	InferredType Analysis[Type]
}

func (s *State) Name() *ast.Identifier {
	return s.AST.Names[s.Index]
}

func (s *State) Value() *ast.Expression {
	if s.Index >= len(s.AST.Values) {
		return nil
	}
	return s.AST.Values[s.Index]
}

func (s *State) ResolvedType() Analysis[Type] {
	if s.AST.Type != nil {
		var a Analysis[Type]
		a.SetResult(Type(s.AST.Type.Type))
		return a
	}
	return s.InferredType
}
