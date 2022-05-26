package lex

// Item represents a lexical item as emitted by the lexer.
type Item struct {
	// Type is the type of the item.
	Type ItemType
	// Value is the value of the item, if any.
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

const (
	Error ItemType = iota
	EOF

	Indent // indention level increased
	Dedent // indention level decreased

	Ident   // identifier
	Literal // after a '.', '#' etc.
	Code    // after a minus, if, etc.
	Text    // the text after an element, pipe etc.

	Comment // '//'

	Import // 'import'
	Func   // 'func'

	Extends // 'extends'
	Include // 'include'

	Doctype // 'doctype'

	Block         // 'block'
	Append        // 'append' or 'block append'
	Prepend       // 'prepend' or 'block prepend'
	BlockIfExists // '?' after 'append' or 'prepend'

	Mixin // 'mixin'

	If     // 'if'
	ElseIf // 'else if'
	Else   // 'else'

	Switch      // 'switch'
	Case        // 'case'
	CaseDefault // 'default

	For   // 'for'
	Range // 'range' keyword used in a for

	LParen   // '('
	RParen   // ')'
	LBrace   // '{'
	RBrace   // '}'
	LBracket // '['
	RBracket // ']'

	And // '&'

	Assign // '=', used in conjunction with attributes or mixin arg defaults
	Comma  // ',' used for mixin args and ternary expressions

	Class     // '.'
	ID        // '#'
	MixinCall // '+'

	CodeAssign     // '=' after an element, e.g. 'p='
	DotBlock       // '.' after an element, e.g. 'p.'
	BlockExpansion // ':' after an element, to indicate block expansion, e.g. 'a: img'

	CodeStart // '-'

	Pipe // '|'

	Hash      // '#'
	Unescaped // '!'

	Ternary // '?' at the start of code

	TagVoid // '/'
)
