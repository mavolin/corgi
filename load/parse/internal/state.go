package parser

import (
	"github.com/mavolin/corgi/v2/file/ast"
)

type State struct {
	ws *State

	index     uint32
	line, col uint16

	numComments uint16
	numErrs     uint8
}

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
func (s *State) NumErrors() uint8 { return s.numErrs }

func (s *State) advance(size uint32, isNL bool) {
	if isNL {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	s.index += size
}

func (s *State) Copy(into *State) {
	*into = *s
}
