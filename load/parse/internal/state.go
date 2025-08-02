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

// commitWS commits the whitespace, preventing rollback.
func (s *State) commitWS() {
	if !s.parsingWS {
		s.ws = nil
	}
}

func (s *State) markWSStart(p *pool[State]) {
	if s.ws != nil {
		return
	}
	s.ws = s.Clone(p)
}

func (s *State) takeWSStart(p *pool[State]) *State {
	if s.ws == nil || s.parsingWS {
		return s.Clone(p)
	}
	wsStart := s.ws
	s.ws = nil
	return wsStart
}

func (s *State) Clone(p *pool[State]) *State {
	s2 := p.Get()
	s2.line = s.line
	s2.col = s.col
	s2.index = s.index
	s2.errs = s.errs
	s2.comments = s.comments
	s2.ws = s.ws
	s2.inline = s.inline
	s2.parsingWS = s.parsingWS
	return s2
}
