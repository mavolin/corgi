package parser

import (
	"github.com/mavolin/corgi/v2/file/ast"
)

type (
	State struct {
		ws *State

		index     uint32
		line, col uint16

		commentLen uint16
		errLen     uint8

		inline    bool
		parsingWS bool
	}
)

func newState() *State {
	return &State{
		line:  1,
		col:   1,
		index: 0,
	}
}

func (s *State) Pos() ast.Position {
	return ast.Position{Line: int(s.line), Col: int(s.col)}
}

func (s *State) Index() int       { return int(s.index) }
func (s *State) NumErrors() uint8 { return s.errLen }

func (s *State) advance(size uint32, isNL bool) {
	if isNL {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	s.index += size
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
	s2.errLen = s.errLen
	s2.commentLen = s.commentLen
	s2.ws = s.ws
	s2.inline = s.inline
	s2.parsingWS = s.parsingWS
	return s2
}
