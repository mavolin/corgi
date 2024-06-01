package parser

import (
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
)

type (
	State struct {
		line, col int
		index     int

		errs     []*fileerr.Error
		comments []*ast.CommentGroup

		inline bool
	}
)

func newState() *State {
	return &State{
		line:     1,
		col:      1,
		index:    0,
		errs:     make([]*fileerr.Error, 0, 48),
		comments: make([]*ast.CommentGroup, 0, 128),
	}
}

func (s *State) Pos() ast.Position {
	return ast.Position{Line: s.line, Col: s.col}
}

func (s *State) advance(size int, isNL bool) {
	if isNL {
		s.line++
		s.col = 1
	} else {
		s.col += size
	}
	s.index += size
}

func (s *State) Errors() []*fileerr.Error {
	return s.errs
}

func (s *State) Comments() []*ast.CommentGroup {
	return s.comments
}

func (s *State) CaptureError(err *fileerr.Error) {
	s.errs = append(s.errs, err)
}

func (s *State) CaptureComment(cg *ast.CommentGroup) {
	s.comments = append(s.comments, cg)
}

func (s State) Clone() *State {
	s2 := s
	return &s2
}
