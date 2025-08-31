package file

import "github.com/mavolin/corgi/v2/file/ast"

type State struct {
	//
	// BUILD SYMBOLS

	AST  *ast.StateSpec
	File *File
	// Index of the variable in the Names and Values slices.
	Index int

	//
	// ANALYZER

	// Analyzed indicates whether the State has been analyzed,
	// albeit with errors.
	Analyzed bool

	// The InferredType of this value, if there is no explicit type.
	InferredType Analysis[string]
}

func (s *State) Name() *ast.Identifier {
	return s.AST.Names[s.Index]
}

func (s *State) Value() *ast.Expression {
	if len(s.AST.Values) == 1 {
		return s.AST.Values[0]
	}
	return s.AST.Values[s.Index]
}

func (s *State) ResolvedType() Analysis[string] {
	if s.AST.Type != nil {
		var a Analysis[string]
		a.SetResult(s.AST.Type.Type)
		return a
	}
	return s.InferredType
}
