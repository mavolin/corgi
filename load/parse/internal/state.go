package parser

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

type (
	State struct {
		line, col int
		index     int

		errs     diagnostic.List
		comments []*ast.CommentGroup

		ws *State

		inline    bool
		parsingWS bool
	}
)

func newState() *State {
	return &State{
		line:     1,
		col:      1,
		index:    0,
		errs:     make(diagnostic.List, 0, 48),
		comments: make([]*ast.CommentGroup, 0, 128),
	}
}

func (s *State) Pos() ast.Position {
	return ast.Position{Line: s.line, Col: s.col}
}

func (s *State) Index() int {
	return s.index
}

func (s *State) advance(size int, isNL bool) {
	if isNL {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	s.index += size
}

func (s *State) Errors() diagnostic.List {
	return s.errs
}

func (s *State) Comments() []*ast.CommentGroup {
	return s.comments
}

func (s *State) CaptureError(err *diagnostic.Diagnostic) {
	s.errs = append(s.errs, err)
}

func (s *State) CaptureComment(cg *ast.CommentGroup) {
	s.comments = append(s.comments, cg)
}

func (s *State) markWSStart() {
	if s.ws != nil {
		return
	}
	s.ws = s.Clone()
}

// commitWS commits the whitespace, preventing rollback.
func (s *State) commitWS() {
	s.ws = nil
}

func (s *State) takeWSStart() *State {
	if s.ws == nil || s.parsingWS {
		return s.Clone()
	}
	wsStart := s.ws
	s.commitWS()
	return wsStart
}

func (s State) Clone() *State {
	s2 := s
	return &s2
}
