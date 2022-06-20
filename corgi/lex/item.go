package lex

// Item represents a lexical item as emitted by the lexer.
type Item struct {
	// Type is the type of the item.
	Type ItemType
	// Expression is the value of the item, if any.
	Val string
	// Err is the error.
	// It is only set when Type is Error.
	Err error
	// Line is the line where the item starts.
	Line int
	// Col is the column after which the item starts.
	Col int
}

// ItemType is an enum representing the type of item the lexer encounters.
//
// The value assigned to an item may change, e.g. if the list ever gets
// expanded.
// Only use ItemTypes in conjunction with their constants.
type ItemType uint8

//go:generate stringer -type ItemType

const (
	Error ItemType = iota
	EOF

	Indent // indention level increased
	Dedent // indention level decreased

	Element    // element name
	Ident      // identifier
	Literal    // after a '.', '#' etc.
	Expression // a Go expression
	Text       // the text after an element, pipe etc. that needs HTML escaping

	CodeStart // '-'
	Code      // after CodeStart

	Ternary     // '?' at the start of code
	TernaryElse // ':' after the ifTrue

	NilCheck // '?' for nil and out-of-bounds checks

	LParen   // '('
	RParen   // ')'
	LBrace   // '{'
	RBrace   // '}'
	LBracket // '['
	RBracket // ']'

	Assign         // '='
	AssignNoEscape // '!='
	Comma          // ',' used for mixin args and ternary expressions

	Comment // '//'

	Import // 'import'
	Func   // 'func'

	Extend  // 'extend'
	Include // 'include'
	Use     // 'use'

	Doctype // 'doctype'

	Block        // 'block'
	BlockAppend  // 'append' or 'block append'
	BlockPrepend // 'prepend' or 'block prepend'

	If      // 'if'
	IfBlock // 'if block'
	ElseIf  // 'else if'
	Else    // 'else'

	Switch      // 'switch'
	Case        // 'case'
	DefaultCase // 'default

	For   // 'for'
	Range // 'range' keyword used in a for

	While // 'while'

	Mixin              // 'mixin'
	MixinCall          // '+'
	MixinBlockShortcut // '>' after a mixin call with a single block

	And // '&'

	Div // sent before a '.' or '#' to indicate a div is being created

	Class // '.'
	ID    // '#'

	BlockExpansion // ':'

	Filter // ':'

	DotBlock     // '.' e.g., after an element, such as 'p.'
	DotBlockLine // used at the start of each line in a DotBlock

	Pipe // '|'

	Hash     // '#'
	NoEscape // '!' after a hash

	TagVoid // '/'
)
