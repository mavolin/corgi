// Code generated by "stringer -type ItemType"; DO NOT EDIT.

package lex

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
	_ = x[Doctype-29]
	_ = x[Block-30]
	_ = x[BlockAppend-31]
	_ = x[BlockPrepend-32]
	_ = x[If-33]
	_ = x[IfBlock-34]
	_ = x[ElseIf-35]
	_ = x[Else-36]
	_ = x[Switch-37]
	_ = x[Case-38]
	_ = x[DefaultCase-39]
	_ = x[For-40]
	_ = x[Range-41]
	_ = x[While-42]
	_ = x[Mixin-43]
	_ = x[MixinCall-44]
	_ = x[MixinBlockShortcut-45]
	_ = x[And-46]
	_ = x[Div-47]
	_ = x[Class-48]
	_ = x[ID-49]
	_ = x[BlockExpansion-50]
	_ = x[Filter-51]
	_ = x[DotBlock-52]
	_ = x[DotBlockLine-53]
	_ = x[Pipe-54]
	_ = x[Hash-55]
	_ = x[NoEscape-56]
	_ = x[TagVoid-57]
}

const _ItemType_name = "ErrorEOFIndentDedentElementIdentLiteralExpressionTextCodeStartCodeTernaryTernaryElseNilCheckLParenRParenLBraceRBraceLBracketRBracketAssignAssignNoEscapeCommaCommentImportFuncExtendIncludeUseDoctypeBlockBlockAppendBlockPrependIfIfBlockElseIfElseSwitchCaseDefaultCaseForRangeWhileMixinMixinCallMixinBlockShortcutAndDivClassIDBlockExpansionFilterDotBlockDotBlockLinePipeHashNoEscapeTagVoid"

var _ItemType_index = [...]uint16{0, 5, 8, 14, 20, 27, 32, 39, 49, 53, 62, 66, 73, 84, 92, 98, 104, 110, 116, 124, 132, 138, 152, 157, 164, 170, 174, 180, 187, 190, 197, 202, 213, 225, 227, 234, 240, 244, 250, 254, 265, 268, 273, 278, 283, 292, 310, 313, 316, 321, 323, 337, 343, 351, 363, 367, 371, 379, 386}

func (i ItemType) String() string {
	if i >= ItemType(len(_ItemType_index)-1) {
		return "ItemType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ItemType_name[_ItemType_index[i]:_ItemType_index[i+1]]
}
