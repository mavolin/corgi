package parser

import (
	"github.com/mavolin/corgi/v2/file/ast"
)

type (
	State struct {
		ws *State

		byteIndex ByteIndex
		runeIndex RuneIndex
		line, col uint32

		numComments uint16
		numErrs     uint8
	}

	ByteIndex uint32
	RuneIndex uint32
)

func newState(startLine, startCol uint32) *State {
	return &State{line: startLine, col: startCol}
}

func (s *State) Pos() ast.Position {
	return ast.Position{Line: int(s.line), Col: int(s.col)}
}

func (s *State) advance(size ByteIndex, isNL bool) {
	if isNL {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
	s.byteIndex += size
	s.runeIndex++
}

func (s *State) Copy(into *State) {
	*into = *s
}
