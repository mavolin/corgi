// Code generated by "stringer -type Token"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Error-0]
	_ = x[EOF-1]
	_ = x[Indent-2]
	_ = x[Dedent-3]
	_ = x[Element-4]
	_ = x[Ident-5]
	_ = x[Literal-6]
	_ = x[Expression-7]
	_ = x[Text-8]
	_ = x[CodeStart-9]
	_ = x[Code-10]
	_ = x[Ternary-11]
	_ = x[TernaryElse-12]
	_ = x[NilCheck-13]
	_ = x[LParen-14]
	_ = x[RParen-15]
	_ = x[LBrace-16]
	_ = x[RBrace-17]
	_ = x[LBracket-18]
	_ = x[RBracket-19]
	_ = x[Assign-20]
	_ = x[AssignNoEscape-21]
	_ = x[Comma-22]
	_ = x[Comment-23]
	_ = x[Import-24]
	_ = x[Func-25]
	_ = x[Extend-26]
	_ = x[Include-27]
	_ = x[Use-28]
	_ = x[Block-29]
	_ = x[Append-30]
	_ = x[Prepend-31]
	_ = x[If-32]
	_ = x[IfBlock-33]
	_ = x[ElseIf-34]
	_ = x[Else-35]
	_ = x[Switch-36]
	_ = x[Case-37]
	_ = x[Default-38]
	_ = x[For-39]
	_ = x[Range-40]
	_ = x[While-41]
	_ = x[Mixin-42]
	_ = x[MixinCall-43]
	_ = x[MixinBlockShorthand-44]
	_ = x[And-45]
	_ = x[Div-46]
	_ = x[Class-47]
	_ = x[ID-48]
	_ = x[BlockExpansion-49]
	_ = x[Filter-50]
	_ = x[DotBlock-51]
	_ = x[DotBlockLine-52]
	_ = x[Pipe-53]
	_ = x[Hash-54]
	_ = x[NoEscape-55]
	_ = x[TagVoid-56]
}

const _Token_name = "ErrorEOFIndentDedentElementIdentLiteralExpressionTextCodeStartCodeTernaryTernaryElseNilCheckLParenRParenLBraceRBraceLBracketRBracketAssignAssignNoEscapeCommaCommentImportFuncExtendIncludeUseBlockAppendPrependIfIfBlockElseIfElseSwitchCaseDefaultForRangeWhileMixinMixinCallMixinBlockShorthandAndDivClassIDBlockExpansionFilterDotBlockDotBlockLinePipeHashNoEscapeTagVoid"

var _Token_index = [...]uint16{0, 5, 8, 14, 20, 27, 32, 39, 49, 53, 62, 66, 73, 84, 92, 98, 104, 110, 116, 124, 132, 138, 152, 157, 164, 170, 174, 180, 187, 190, 195, 201, 208, 210, 217, 223, 227, 233, 237, 244, 247, 252, 257, 262, 271, 290, 293, 296, 301, 303, 317, 323, 331, 343, 347, 351, 359, 366}

func (i Token) String() string {
	if i >= Token(len(_Token_index)-1) {
		return "Token(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Token_name[_Token_index[i]:_Token_index[i+1]]
}