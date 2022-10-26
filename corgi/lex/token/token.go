package token

// Token is an enum representing the type of item the lexer encounters.
//
// The value assigned to an item may change, e.g. if the list ever gets
// expanded.
// Only use Tokens in conjunction with their constants.
type Token uint8

//go:generate stringer -type Token

const (
	Error Token = iota
	EOF

	Indent // indention level increased
	Dedent // indention level decreased

	Element // element name
	// Ident is an identifier.
	//
	// It starts with a unicode letter or underscore.
	// It is followed by any number of unicode letters, decimal digits, or
	// underscores.
	// This is the same pattern as Go uses for its identifiers.
	Ident
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

	Block   // 'block'
	Append  // 'append'
	Prepend // 'prepend'

	If      // 'if'
	IfBlock // 'if block'
	ElseIf  // 'else if'
	Else    // 'else'

	Switch  // 'switch'
	Case    // 'case'
	Default // 'default

	For   // 'for'
	Range // 'range' keyword used in a for

	While // 'while'

	Mixin               // 'mixin'
	MixinCall           // '+'
	MixinBlockShorthand // '>'

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